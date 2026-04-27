package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/contracts"
	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/platforms"
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

func TestExecuteCIGenerateRejectsUnsafeGeneratedPath(t *testing.T) {
	restore := registerTestGenerator(t, "unsafe", staticGenerator{
		files: map[string]string{"../escape.yml": "bad"},
	})
	defer restore()

	result, err := ExecuteCIGenerate(context.Background(), sdk.TypedStepRequest[*contracts.CIGenerateConfig, *contracts.CIGenerateInput]{
		Config: &contracts.CIGenerateConfig{},
		Input: &contracts.CIGenerateInput{
			Platform:  "unsafe",
			OutputDir: t.TempDir(),
		},
	})
	if err != nil {
		t.Fatalf("ExecuteCIGenerate: %v", err)
	}
	if result == nil || result.Output == nil || result.Output.Error == "" {
		t.Fatalf("expected unsafe path error, got %#v", result)
	}
}

func TestExecuteCIGenerateSortsFilesWritten(t *testing.T) {
	outputDir := t.TempDir()
	restore := registerTestGenerator(t, "static", staticGenerator{
		files: map[string]string{
			"z.yml": "z",
			"a.yml": "a",
		},
	})
	defer restore()

	result, err := ExecuteCIGenerate(context.Background(), sdk.TypedStepRequest[*contracts.CIGenerateConfig, *contracts.CIGenerateInput]{
		Config: &contracts.CIGenerateConfig{},
		Input: &contracts.CIGenerateInput{
			Platform:  "static",
			OutputDir: outputDir,
		},
	})
	if err != nil {
		t.Fatalf("ExecuteCIGenerate: %v", err)
	}
	want := []string{
		filepath.Join(outputDir, "a.yml"),
		filepath.Join(outputDir, "z.yml"),
	}
	if len(result.Output.FilesWritten) != len(want) {
		t.Fatalf("FilesWritten length = %d, want %d", len(result.Output.FilesWritten), len(want))
	}
	for i := range want {
		if result.Output.FilesWritten[i] != want[i] {
			t.Fatalf("FilesWritten[%d] = %q, want %q", i, result.Output.FilesWritten[i], want[i])
		}
	}
}

type staticGenerator struct {
	files map[string]string
}

func (g staticGenerator) Generate(_ platforms.Options) (map[string]string, error) {
	return g.files, nil
}

func registerTestGenerator(t *testing.T, platform string, generator Generator) func() {
	t.Helper()
	original, existed := registry[platform]
	registry[platform] = func() Generator { return generator }
	return func() {
		if existed {
			registry[platform] = original
			return
		}
		delete(registry, platform)
	}
}
