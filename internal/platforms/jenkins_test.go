package platforms_test

import (
	"strings"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/platforms"
)

func TestJenkinsGenerator_Generate(t *testing.T) {
	g := platforms.NewJenkinsGenerator()
	opts := platforms.Options{
		InfraConfig:   "infra.yaml",
		ProjectName:   "my-project",
		DefaultBranch: "main",
	}

	files, err := g.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if _, ok := files["Jenkinsfile"]; !ok {
		t.Fatal("expected Jenkinsfile in output")
	}
}

func TestJenkinsGenerator_DeclarativeSyntax(t *testing.T) {
	g := platforms.NewJenkinsGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "main"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files["Jenkinsfile"]

	// Must use declarative pipeline syntax
	if !strings.HasPrefix(strings.TrimSpace(content), "pipeline {") {
		t.Errorf("Jenkinsfile must start with 'pipeline {', got:\n%s", content[:min(80, len(content))])
	}

	// Must have agent block
	if !strings.Contains(content, "agent {") {
		t.Error("Jenkinsfile must have 'agent {' block")
	}

	// Must have stages block
	if !strings.Contains(content, "stages {") {
		t.Error("Jenkinsfile must have 'stages {' block")
	}

	// Each stage must use 'steps {' not raw sh calls
	if !strings.Contains(content, "steps {") {
		t.Error("Jenkinsfile must have 'steps {' inside stages")
	}
}

func TestJenkinsGenerator_RequiredStages(t *testing.T) {
	g := platforms.NewJenkinsGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "main", InfraConfig: "infra.yaml"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files["Jenkinsfile"]
	for _, stage := range []string{"Plan", "Build", "Apply", "Deploy"} {
		if !strings.Contains(content, "stage('"+stage+"')") {
			t.Errorf("expected stage('%s') in Jenkinsfile", stage)
		}
	}
}

func TestJenkinsGenerator_WfctlCommands(t *testing.T) {
	g := platforms.NewJenkinsGenerator()
	files, err := g.Generate(platforms.Options{InfraConfig: "infra.yaml", DefaultBranch: "main"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files["Jenkinsfile"]
	checks := []string{
		"wfctl infra plan",
		"wfctl infra apply",
		"--auto-approve",
		"wfctl deploy",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q in Jenkinsfile\ngot:\n%s", want, content)
		}
	}
}

func TestJenkinsGenerator_DefaultBranch(t *testing.T) {
	g := platforms.NewJenkinsGenerator()
	files, err := g.Generate(platforms.Options{DefaultBranch: "trunk"})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	content := files["Jenkinsfile"]
	if !strings.Contains(content, "trunk") {
		t.Error("expected custom default branch 'trunk' in Jenkinsfile")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
