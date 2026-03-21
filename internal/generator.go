package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/platforms"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// Platform constants.
const (
	PlatformGitHubActions = "github_actions"
	PlatformGitLabCI      = "gitlab_ci"
	PlatformJenkins       = "jenkins"
	PlatformCircleCI      = "circleci"
)

// Generator defines the interface all platform generators implement.
type Generator interface {
	// Generate produces CI config files. Returns a map of relative output path → content.
	Generate(opts platforms.Options) (map[string]string, error)
}

// registry maps platform names to generator constructors.
var registry = map[string]func() Generator{
	PlatformGitHubActions: func() Generator { return platforms.NewGitHubActionsGenerator() },
	PlatformGitLabCI:      func() Generator { return platforms.NewGitLabCIGenerator() },
	PlatformJenkins:       func() Generator { return platforms.NewJenkinsGenerator() },
	PlatformCircleCI:      func() Generator { return platforms.NewCircleCIGenerator() },
}

// ciGenerateStep implements step.ci_generate.
type ciGenerateStep struct {
	name string
}

func newCIGenerateStep(name string, _ map[string]any) (*ciGenerateStep, error) {
	return &ciGenerateStep{name: name}, nil
}

// Execute generates CI/CD config files for the specified platform.
//
// Config keys:
//   - platform (string, required): github_actions | gitlab_ci | jenkins | circleci
//   - output_dir (string, required): directory to write generated files into
//   - infra_config (string): path to infra.yaml (default: "infra.yaml")
//   - project_name (string): project/repo name (default: "my-project")
//   - runner (string): runner label for GHA (default: "self-hosted, Linux, X64")
//   - default_branch (string): main branch name (default: "main")
func (s *ciGenerateStep) Execute(ctx context.Context, _ map[string]any, _ map[string]map[string]any, current map[string]any, _ map[string]any, config map[string]any) (*sdk.StepResult, error) {
	platform := resolveString("platform", current, config)
	if platform == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "platform is required"}}, nil
	}

	outputDir := resolveString("output_dir", current, config)
	if outputDir == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "output_dir is required"}}, nil
	}

	newGen, ok := registry[platform]
	if !ok {
		return &sdk.StepResult{Output: map[string]any{"error": fmt.Sprintf("unknown platform %q", platform)}}, nil
	}

	opts := platforms.Options{
		InfraConfig:   resolveStringDefault("infra_config", current, config, "infra.yaml"),
		ProjectName:   resolveStringDefault("project_name", current, config, "my-project"),
		Runner:        resolveStringDefault("runner", current, config, "self-hosted, Linux, X64"),
		DefaultBranch: resolveStringDefault("default_branch", current, config, "main"),
	}

	gen := newGen()
	files, err := gen.Generate(opts)
	if err != nil {
		return &sdk.StepResult{Output: map[string]any{"error": err.Error()}}, nil
	}

	written := make([]string, 0, len(files))
	for relPath, content := range files {
		fullPath := filepath.Join(outputDir, relPath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return &sdk.StepResult{Output: map[string]any{"error": fmt.Sprintf("mkdir %s: %v", dir, err)}}, nil
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return &sdk.StepResult{Output: map[string]any{"error": fmt.Sprintf("write %s: %v", fullPath, err)}}, nil
		}
		written = append(written, fullPath)
	}

	return &sdk.StepResult{Output: map[string]any{
		"platform":      platform,
		"output_dir":    outputDir,
		"files_written": written,
		"file_count":    len(written),
	}}, nil
}

// resolveString reads a string value from current (runtime) first, then config.
func resolveString(key string, current, config map[string]any) string {
	if v, ok := current[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// resolveStringDefault is resolveString with a fallback default.
func resolveStringDefault(key string, current, config map[string]any, def string) string {
	if v := resolveString(key, current, config); v != "" {
		return v
	}
	return def
}
