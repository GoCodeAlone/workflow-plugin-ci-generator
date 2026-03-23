package cigenerator_test

import (
	"context"
	"testing"

	"github.com/GoCodeAlone/workflow/wftest"
)

// TestIntegration_GenerateGitHubActions verifies that a pipeline using
// step.ci_generate executes and returns GitHub Actions output.
func TestIntegration_GenerateGitHubActions(t *testing.T) {
	h := wftest.New(t,
		wftest.WithYAML(`
pipelines:
  generate-github:
    steps:
      - name: generate
        type: step.ci_generate
        config:
          platform: github-actions
          output_path: /tmp/ci.yml
`),
		wftest.MockStep("step.ci_generate", wftest.Returns(map[string]any{
			"output_path": "/tmp/ci.yml",
			"platform":    "github-actions",
			"success":     true,
		})),
	)

	result := h.ExecutePipeline("generate-github", nil)
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}
	if !result.StepExecuted("generate") {
		t.Fatal("expected step 'generate' to be executed")
	}
	out := result.StepOutput("generate")
	if out["platform"] != "github-actions" {
		t.Errorf("expected platform=github-actions, got %v", out["platform"])
	}
	if out["success"] != true {
		t.Errorf("expected success=true, got %v", out["success"])
	}
}

// TestIntegration_GenerateGitLabCI verifies that a pipeline using
// step.ci_generate returns GitLab CI configuration output.
func TestIntegration_GenerateGitLabCI(t *testing.T) {
	h := wftest.New(t,
		wftest.WithYAML(`
pipelines:
  generate-gitlab:
    steps:
      - name: generate
        type: step.ci_generate
        config:
          platform: gitlab-ci
          output_path: /tmp/.gitlab-ci.yml
`),
		wftest.MockStep("step.ci_generate", wftest.Returns(map[string]any{
			"output_path": "/tmp/.gitlab-ci.yml",
			"platform":    "gitlab-ci",
			"success":     true,
		})),
	)

	result := h.ExecutePipeline("generate-gitlab", nil)
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}
	if !result.StepExecuted("generate") {
		t.Fatal("expected step 'generate' to be executed")
	}
	out := result.StepOutput("generate")
	if out["platform"] != "gitlab-ci" {
		t.Errorf("expected platform=gitlab-ci, got %v", out["platform"])
	}
	if out["output_path"] != "/tmp/.gitlab-ci.yml" {
		t.Errorf("expected output_path=/tmp/.gitlab-ci.yml, got %v", out["output_path"])
	}
}

// TestIntegration_MultiStepPipeline verifies that data flows correctly through
// a pipeline with step.ci_generate followed by a step.set step.
func TestIntegration_MultiStepPipeline(t *testing.T) {
	h := wftest.New(t,
		wftest.WithYAML(`
pipelines:
  multi-step:
    steps:
      - name: generate
        type: step.ci_generate
        config:
          platform: github-actions
          output_path: /tmp/ci.yml
      - name: record
        type: step.set
        config:
          key: processed
          value: "true"
`),
		wftest.MockStep("step.ci_generate", wftest.Returns(map[string]any{
			"output_path": "/tmp/ci.yml",
			"platform":    "github-actions",
			"success":     true,
		})),
		wftest.MockStep("step.set", wftest.Returns(map[string]any{
			"processed": "true",
		})),
	)

	result := h.ExecutePipeline("multi-step", nil)
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}
	if result.StepCount() < 2 {
		t.Errorf("expected at least 2 steps executed, got %d", result.StepCount())
	}
	if !result.StepExecuted("generate") {
		t.Fatal("expected step 'generate' to be executed")
	}
	if !result.StepExecuted("record") {
		t.Fatal("expected step 'record' to be executed")
	}
	// Verify generate output is present
	genOut := result.StepOutput("generate")
	if genOut["success"] != true {
		t.Errorf("expected generate success=true, got %v", genOut["success"])
	}
	// Verify record output is present
	recOut := result.StepOutput("record")
	if recOut["processed"] != "true" {
		t.Errorf("expected record processed=true, got %v", recOut["processed"])
	}
}

// TestIntegration_CIGenerateWithInput verifies that trigger data passed to the
// pipeline is available to the step handler as input.
func TestIntegration_CIGenerateWithInput(t *testing.T) {
	var capturedInput map[string]any

	h := wftest.New(t,
		wftest.WithYAML(`
pipelines:
  generate-with-input:
    steps:
      - name: generate
        type: step.ci_generate
        config:
          platform: circleci
          output_path: /tmp/circleci.yml
`),
		wftest.MockStep("step.ci_generate", wftest.StepHandlerFunc(
			func(ctx context.Context, config, input map[string]any) (map[string]any, error) {
				capturedInput = input
				return map[string]any{
					"output_path": "/tmp/circleci.yml",
					"platform":    "circleci",
					"success":     true,
					"repo":        input["repo"],
				}, nil
			},
		)),
	)

	triggerData := map[string]any{
		"repo":   "my-org/my-repo",
		"branch": "main",
	}
	result := h.ExecutePipeline("generate-with-input", triggerData)
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}
	if !result.StepExecuted("generate") {
		t.Fatal("expected step 'generate' to be executed")
	}
	// Verify trigger data reached the step
	if capturedInput == nil {
		t.Fatal("expected step to receive input")
	}
	if capturedInput["repo"] != "my-org/my-repo" {
		t.Errorf("expected repo=my-org/my-repo in step input, got %v", capturedInput["repo"])
	}
	// Verify output reflects input data
	out := result.StepOutput("generate")
	if out["repo"] != "my-org/my-repo" {
		t.Errorf("expected output repo=my-org/my-repo, got %v", out["repo"])
	}
}
