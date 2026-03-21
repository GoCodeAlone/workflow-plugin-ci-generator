// Package internal implements the workflow-plugin-ci-generator plugin.
package internal

import (
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// ciGeneratorPlugin implements sdk.PluginProvider and sdk.StepProvider.
type ciGeneratorPlugin struct{}

// NewCIGeneratorPlugin returns a new ciGeneratorPlugin instance.
func NewCIGeneratorPlugin() sdk.PluginProvider {
	return &ciGeneratorPlugin{}
}

// Manifest returns plugin metadata.
func (p *ciGeneratorPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "workflow-plugin-ci-generator",
		Version:     "0.1.0",
		Author:      "GoCodeAlone",
		Description: "CI/CD config generator for GitHub Actions, GitLab CI, Jenkins, and CircleCI",
	}
}

// StepTypes returns the step type names this plugin provides.
func (p *ciGeneratorPlugin) StepTypes() []string {
	return []string{
		"step.ci_generate",
	}
}

// CreateStep creates a step instance of the given type.
func (p *ciGeneratorPlugin) CreateStep(typeName, name string, config map[string]any) (sdk.StepInstance, error) {
	switch typeName {
	case "step.ci_generate":
		return newCIGenerateStep(name, config)
	default:
		return nil, fmt.Errorf("ci-generator plugin: unknown step type %q", typeName)
	}
}
