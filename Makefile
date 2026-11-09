.PHONY: check hooks build

check:
	@git ls-files -z "*.go" | xargs -0 gofmt -l | (! grep .)
	go vet ./...
	go test ./...
	GOOS=windows go vet ./...
	GOOS=linux go vet ./...
	npm test
	bash scripts/check-fixture-pairs.sh

hooks:
	go tool lefthook install

build:
	go build -o bin/review-go ./cmd/review

.PHONY: bench
bench:
	go run ./cmd/bench

# Validate .goreleaser.yaml with the version pinned in the release workflow.
.PHONY: release-check
release-check:
	go run github.com/goreleaser/goreleaser/v2@v2.20.0 check
