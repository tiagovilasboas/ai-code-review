package source

import (
	"os"
	"strings"
)

// repoEnv names the variables that point git at a repository other than the
// one in cmd.Dir. A git hook exports GIT_DIR and friends, so a review run from
// a hook (or this repo's own tests under lefthook) would otherwise read, and
// for the tests write, the hook's repository instead of the requested one.
var repoEnv = []string{
	"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR",
	"GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES",
	"GIT_NAMESPACE", "GIT_PREFIX",
}

// gitEnv returns the process environment without repoEnv, so git locates the
// repository from its working directory alone.
func gitEnv() []string {
	env := os.Environ()
	out := env[:0:0]
	for _, kv := range env {
		name, _, _ := strings.Cut(kv, "=")
		if !isRepoVar(name) {
			out = append(out, kv)
		}
	}
	return out
}

func isRepoVar(name string) bool {
	for _, v := range repoEnv {
		if strings.EqualFold(name, v) {
			return true
		}
	}
	return false
}
