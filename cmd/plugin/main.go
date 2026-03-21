// Command workflow-plugin-ci-generator is a workflow engine external plugin that
// generates CI/CD config files for GitHub Actions, GitLab CI, Jenkins, and CircleCI.
// It runs as a subprocess and communicates with the host workflow engine via
// the go-plugin protocol.
package main

import (
	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func main() {
	sdk.Serve(internal.NewCIGeneratorPlugin())
}
