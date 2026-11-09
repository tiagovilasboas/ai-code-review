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
