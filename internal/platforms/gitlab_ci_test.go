package platforms_test

import (
	"strings"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/platforms"
)

func TestGitLabCIGenerator_Generate(t *testing.T) {
	g := platforms.NewGitLabCIGenerator()
	opts := platforms.Options{
		InfraConfig:   "infra.yaml",
		ProjectName:   "my-project",
		DefaultBranch: "main",
	}

	files, err := g.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if _, ok := files[".gitlab-ci.yml"]; !ok {
		t.Fatal("expected .gitlab-ci.yml in output")
	}
}

func TestGitLabCIGenerator_Syntax(t *testing.T) {
	g := platforms.NewGitLabCIGenerator()
	opts := platforms.Options{
		InfraConfig:   "infra.yaml",
		DefaultBranch: "main",
	}

	files, err := g.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".gitlab-ci.yml"]

	// Must use rules: not only:
	if strings.Contains(content, "\nonly:") {
		t.Error(".gitlab-ci.yml must not use deprecated 'only:' syntax")
	}
	if !strings.Contains(content, "rules:") {
		t.Error(".gitlab-ci.yml must use 'rules:' syntax")
	}

	// Must use needs: for DAG
	if !strings.Contains(content, "needs:") {
		t.Error(".gitlab-ci.yml must use 'needs:' for DAG pipeline ordering")
	}

	// Must have environment: for deployment tracking
	if !strings.Contains(content, "environment:") {
		t.Error(".gitlab-ci.yml must use 'environment:' for deployment tracking")
	}
}

func TestGitLabCIGenerator_Stages(t *testing.T) {
	g := platforms.NewGitLabCIGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "main"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".gitlab-ci.yml"]
	for _, stage := range []string{"plan", "apply", "build", "deploy"} {
		if !strings.Contains(content, "  - "+stage) {
			t.Errorf("expected stage %q in stages list", stage)
		}
	}
}

func TestGitLabCIGenerator_WfctlCommands(t *testing.T) {
	g := platforms.NewGitLabCIGenerator()
	files, err := g.Generate(platforms.Options{InfraConfig: "infra.yaml", DefaultBranch: "main"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".gitlab-ci.yml"]
	checks := []string{
		"wfctl infra plan",
		"wfctl infra apply",
		"--auto-approve",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q in .gitlab-ci.yml\ngot:\n%s", want, content)
		}
	}
}

func TestGitLabCIGenerator_DefaultBranch(t *testing.T) {
	g := platforms.NewGitLabCIGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "develop"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".gitlab-ci.yml"]
	if !strings.Contains(content, "develop") {
		t.Error("expected custom default branch 'develop' in .gitlab-ci.yml")
	}
}
