package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/contracts"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func TestExecuteCIGenerateTyped(t *testing.T) {
	outputDir := t.TempDir()

	result, err := ExecuteCIGenerate(context.Background(), sdk.TypedStepRequest[*contracts.CIGenerateConfig, *contracts.CIGenerateInput]{
		Config: &contracts.CIGenerateConfig{
			ProjectName:   "typed-project",
			DefaultBranch: "main",
		},
		Input: &contracts.CIGenerateInput{
			Platform:  PlatformGitHubActions,
			OutputDir: outputDir,
		},
	})
	if err != nil {
		t.Fatalf("ExecuteCIGenerate: %v", err)
	}
	if result == nil || result.Output == nil {
		t.Fatal("expected typed output")
	}
	if result.Output.Error != "" {
		t.Fatalf("unexpected output error: %s", result.Output.Error)
	}
	if result.Output.Platform != PlatformGitHubActions {
		t.Fatalf("expected platform %s, got %s", PlatformGitHubActions, result.Output.Platform)
	}
	if result.Output.FileCount == 0 {
		t.Fatal("expected generated files")
	}
	for _, written := range result.Output.FilesWritten {
		if _, err := os.Stat(written); err != nil {
			t.Fatalf("expected generated file %s: %v", written, err)
		}
		if !filepath.IsAbs(written) {
			t.Fatalf("expected absolute generated file path, got %s", written)
		}
	}
}

func TestExecuteCIGenerateTypedValidation(t *testing.T) {
	result, err := ExecuteCIGenerate(context.Background(), sdk.TypedStepRequest[*contracts.CIGenerateConfig, *contracts.CIGenerateInput]{
		Config: &contracts.CIGenerateConfig{},
		Input:  &contracts.CIGenerateInput{},
	})
	if err != nil {
		t.Fatalf("ExecuteCIGenerate: %v", err)
	}
	if result == nil || result.Output == nil || result.Output.Error != "platform is required" {
		t.Fatalf("expected platform validation error, got %#v", result)
	}
}
