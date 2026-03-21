package platforms_test

import (
	"strings"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/platforms"
)

func TestCircleCIGenerator_Generate(t *testing.T) {
	g := platforms.NewCircleCIGenerator()
	opts := platforms.Options{
		InfraConfig:   "infra.yaml",
		ProjectName:   "my-project",
		DefaultBranch: "main",
	}

	files, err := g.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if _, ok := files[".circleci/config.yml"]; !ok {
		t.Fatal("expected .circleci/config.yml in output")
	}
}

func TestCircleCIGenerator_Version(t *testing.T) {
	g := platforms.NewCircleCIGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "main"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".circleci/config.yml"]
	if !strings.HasPrefix(strings.TrimSpace(content), "version: 2.1") {
		t.Errorf("config.yml must start with 'version: 2.1', got:\n%s", content[:min(80, len(content))])
	}
}

func TestCircleCIGenerator_OrbsAndExecutors(t *testing.T) {
	g := platforms.NewCircleCIGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "main"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".circleci/config.yml"]

	if !strings.Contains(content, "orbs:") {
		t.Error("config.yml must have 'orbs:' section")
	}
	if !strings.Contains(content, "executors:") {
		t.Error("config.yml must have 'executors:' section")
	}
}

func TestCircleCIGenerator_ApprovalJob(t *testing.T) {
	g := platforms.NewCircleCIGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "main"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".circleci/config.yml"]
	if !strings.Contains(content, "type: approval") {
		t.Error("config.yml must include an approval job for apply")
	}
}

func TestCircleCIGenerator_WfctlCommands(t *testing.T) {
	g := platforms.NewCircleCIGenerator()
	files, err := g.Generate(platforms.Options{InfraConfig: "infra.yaml", DefaultBranch: "main"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".circleci/config.yml"]
	checks := []string{
		"wfctl infra plan",
		"wfctl infra apply",
		"--auto-approve",
		"wfctl deploy",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q in .circleci/config.yml\ngot:\n%s", want, content)
		}
	}
}

func TestCircleCIGenerator_WorkflowsSection(t *testing.T) {
	g := platforms.NewCircleCIGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "main"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".circleci/config.yml"]
	if !strings.Contains(content, "workflows:") {
		t.Error("config.yml must have 'workflows:' section")
	}
}

func TestCircleCIGenerator_DefaultBranch(t *testing.T) {
	g := platforms.NewCircleCIGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "release"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files[".circleci/config.yml"]
	if strings.Count(content, "release") < 2 {
		t.Error("expected custom branch 'release' to appear multiple times in config.yml")
	}
}
