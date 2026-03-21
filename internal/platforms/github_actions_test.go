package platforms_test

import (
	"strings"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/platforms"
)

func TestGitHubActionsGenerator_Generate(t *testing.T) {
	g := platforms.NewGitHubActionsGenerator()
	opts := platforms.Options{
		InfraConfig:   "infra.yaml",
		ProjectName:   "my-project",
		Runner:        "self-hosted, Linux, X64",
		DefaultBranch: "main",
	}

	files, err := g.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	expectedFiles := []string{
		".github/workflows/infra.yml",
		".github/workflows/build.yml",
		".github/workflows/deploy.yml",
	}
	for _, f := range expectedFiles {
		if _, ok := files[f]; !ok {
			t.Errorf("expected file %q not generated", f)
		}
	}
}

func TestGitHubActionsGenerator_InfraYML(t *testing.T) {
	g := platforms.NewGitHubActionsGenerator()
	opts := platforms.Options{
		InfraConfig:   "infra.yaml",
		DefaultBranch: "main",
		Runner:        "self-hosted, Linux, X64",
	}

	files, err := g.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".github/workflows/infra.yml"]

	checks := []string{
		"name: Infrastructure",
		"uses: actions/checkout@v4",
		"uses: GoCodeAlone/setup-wfctl@v1",
		"permissions:",
		"contents: read",
		"pull-requests: write",
		"uses: actions/github-script@v7",
		"wfctl infra plan -c infra.yaml --output plan.json",
		"wfctl infra plan -c infra.yaml --format markdown > plan.md",
		"wfctl infra apply -c infra.yaml --auto-approve",
		"refs/heads/main",
		"self-hosted, Linux, X64",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("infra.yml: expected to contain %q\ngot:\n%s", want, content)
		}
	}
}

func TestGitHubActionsGenerator_BuildYML(t *testing.T) {
	g := platforms.NewGitHubActionsGenerator()
	opts := platforms.Options{
		DefaultBranch: "main",
		Runner:        "self-hosted, Linux, X64",
	}

	files, err := g.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".github/workflows/build.yml"]

	checks := []string{
		"name: Build",
		"uses: actions/checkout@v4",
		"uses: actions/setup-go@v5",
		"go-version-file: go.mod",
		"go test ./...",
		"go build ./...",
		"docker/login-action@v3",
		"docker/build-push-action@v6",
		"ghcr.io",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("build.yml: expected to contain %q\ngot:\n%s", want, content)
		}
	}
}

func TestGitHubActionsGenerator_DeployYML(t *testing.T) {
	g := platforms.NewGitHubActionsGenerator()
	opts := platforms.Options{
		InfraConfig:   "infra.yaml",
		DefaultBranch: "main",
		Runner:        "self-hosted, Linux, X64",
	}

	files, err := g.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".github/workflows/deploy.yml"]

	checks := []string{
		"name: Deploy",
		"workflow_run:",
		"Infrastructure",
		"wfctl infra apply -c infra.yaml --auto-approve",
		"wfctl deploy --image",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("deploy.yml: expected to contain %q\ngot:\n%s", want, content)
		}
	}
}

func TestGitHubActionsGenerator_DefaultsApplied(t *testing.T) {
	g := platforms.NewGitHubActionsGenerator()
	// empty options — defaults should kick in
	files, err := g.Generate(platforms.Options{})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	infra := files[".github/workflows/infra.yml"]
	if !strings.Contains(infra, "infra.yaml") {
		t.Error("expected default infra config path 'infra.yaml'")
	}
	if !strings.Contains(infra, "refs/heads/main") {
		t.Error("expected default branch 'main'")
	}
}

func TestGitHubActionsGenerator_CustomRunner(t *testing.T) {
	g := platforms.NewGitHubActionsGenerator()
	opts := platforms.Options{
		Runner:        "ubuntu-latest",
		DefaultBranch: "main",
	}
	files, err := g.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if !strings.Contains(files[".github/workflows/infra.yml"], "ubuntu-latest") {
		t.Error("expected custom runner 'ubuntu-latest' in infra.yml")
	}
}
