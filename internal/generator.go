package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/contracts"
	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/platforms"
	"github.com/GoCodeAlone/workflow/cigen"
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

// registry maps platform names to template generator constructors.
// Only jenkins and circleci are handled here; github_actions and gitlab_ci
// are routed through the cigen smart analyzer in ExecuteCIGenerate.
var registry = map[string]func() Generator{
	PlatformJenkins:  func() Generator { return platforms.NewJenkinsGenerator() },
	PlatformCircleCI: func() Generator { return platforms.NewCircleCIGenerator() },
}

// knownPlatforms is the complete set of supported platform names.
var knownPlatforms = map[string]bool{
	PlatformGitHubActions: true,
	PlatformGitLabCI:      true,
	PlatformJenkins:       true,
	PlatformCircleCI:      true,
}

// ExecuteCIGenerate generates CI/CD config files for the specified platform.
// For github_actions and gitlab_ci, the cigen smart analyzer is used
// (analyze → CIPlan → render). For jenkins and circleci, the existing
// template generators are used unchanged.
func ExecuteCIGenerate(ctx context.Context, req sdk.TypedStepRequest[*contracts.CIGenerateConfig, *contracts.CIGenerateInput]) (*sdk.TypedStepResult[*contracts.CIGenerateOutput], error) {
	_ = ctx
	platform := resolveTypedString(req.Input.GetPlatform(), req.Config.GetPlatform())
	if platform == "" {
		return typedCIGenerateError("platform is required"), nil
	}

	outputDir := resolveTypedString(req.Input.GetOutputDir(), req.Config.GetOutputDir())
	if outputDir == "" {
		return typedCIGenerateError("output_dir is required"), nil
	}

	if !knownPlatforms[platform] {
		return typedCIGenerateError(fmt.Sprintf("unknown platform %q", platform)), nil
	}

	infraConfig := resolveTypedStringDefault(req.Input.GetInfraConfig(), req.Config.GetInfraConfig(), "infra.yaml")
	projectName := resolveTypedStringDefault(req.Input.GetProjectName(), req.Config.GetProjectName(), "my-project")
	runner := resolveTypedStringDefault(req.Input.GetRunner(), req.Config.GetRunner(), "self-hosted, Linux, X64")
	defaultBranch := resolveTypedStringDefault(req.Input.GetDefaultBranch(), req.Config.GetDefaultBranch(), "main")
	fromPlan := resolveTypedString(req.Input.GetFromPlan(), req.Config.GetFromPlan())
	phaseConfig := resolveTypedString(req.Input.GetPhaseConfig(), req.Config.GetPhaseConfig())

	var files map[string]string

	switch platform {
	case PlatformGitHubActions, PlatformGitLabCI:
		var plan *cigen.CIPlan
		if fromPlan != "" {
			// Load a pre-computed CIPlan JSON directly.
			raw, err := os.ReadFile(fromPlan)
			if err != nil {
				return typedCIGenerateError(fmt.Sprintf("read from_plan %s: %v", fromPlan, err)), nil
			}
			plan = &cigen.CIPlan{}
			if err := json.Unmarshal(raw, plan); err != nil {
				return typedCIGenerateError(fmt.Sprintf("parse from_plan %s: %v", fromPlan, err)), nil
			}
		} else {
			// Run the smart analyzer.
			opts := cigen.Options{
				Runner:        runner,
				DefaultBranch: defaultBranch,
				Project:       projectName,
				PhaseConfig:   phaseConfig,
			}
			var err error
			plan, err = cigen.Analyze([]string{infraConfig}, opts)
			if err != nil {
				return typedCIGenerateError(fmt.Sprintf("cigen analyze: %v", err)), nil
			}
		}

		var err error
		switch platform {
		case PlatformGitHubActions:
			files, err = cigen.RenderGitHubActions(plan)
		case PlatformGitLabCI:
			files, err = cigen.RenderGitLabCI(plan)
		}
		if err != nil {
			return typedCIGenerateError(fmt.Sprintf("cigen render: %v", err)), nil
		}

	default:
		// jenkins and circleci: template generators.
		opts := platforms.Options{
			InfraConfig:   infraConfig,
			ProjectName:   projectName,
			Runner:        runner,
			DefaultBranch: defaultBranch,
		}
		gen := registry[platform]()
		var err error
		files, err = gen.Generate(opts)
		if err != nil {
			return typedCIGenerateError(err.Error()), nil
		}
	}

	written := make([]string, 0, len(files))
	relPaths := make([]string, 0, len(files))
	for relPath := range files {
		relPaths = append(relPaths, relPath)
	}
	sort.Strings(relPaths)
	for _, relPath := range relPaths {
		if err := validateRelativeOutputPath(relPath); err != nil {
			return typedCIGenerateError(err.Error()), nil
		}
		content := files[relPath]
		fullPath := filepath.Join(outputDir, relPath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return typedCIGenerateError(fmt.Sprintf("mkdir %s: %v", dir, err)), nil
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return typedCIGenerateError(fmt.Sprintf("write %s: %v", fullPath, err)), nil
		}
		written = append(written, fullPath)
	}

	return &sdk.TypedStepResult[*contracts.CIGenerateOutput]{
		Output: &contracts.CIGenerateOutput{
			Platform:     platform,
			OutputDir:    outputDir,
			FilesWritten: written,
			FileCount:    int32(len(written)),
		},
	}, nil
}

func validateRelativeOutputPath(relPath string) error {
	if relPath == "" {
		return fmt.Errorf("generated output path is empty")
	}
	if filepath.IsAbs(relPath) {
		return fmt.Errorf("generated output path %q must be relative", relPath)
	}
	for _, segment := range strings.Split(filepath.ToSlash(relPath), "/") {
		if segment == ".." {
			return fmt.Errorf("generated output path %q must not contain parent directory segments", relPath)
		}
	}
	return nil
}

func typedCIGenerateError(message string) *sdk.TypedStepResult[*contracts.CIGenerateOutput] {
	return &sdk.TypedStepResult[*contracts.CIGenerateOutput]{
		Output: &contracts.CIGenerateOutput{Error: message},
	}
}

func resolveTypedString(input, config string) string {
	if input != "" {
		return input
	}
	return config
}

func resolveTypedStringDefault(input, config, def string) string {
	if value := resolveTypedString(input, config); value != "" {
		return value
	}
	return def
}
