// Package internal implements the workflow-plugin-ci-generator plugin.
package internal

import (
	"fmt"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/contracts"
	pb "github.com/GoCodeAlone/workflow/plugin/external/proto"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
)

// Version is set at build time via -ldflags
// "-X github.com/GoCodeAlone/workflow-plugin-ci-generator/internal.Version=X.Y.Z"
var Version = "dev"

// ciGeneratorPlugin implements sdk.PluginProvider, sdk.TypedStepProvider, and sdk.ContractProvider.
type ciGeneratorPlugin struct{}

// NewCIGeneratorPlugin returns a new ciGeneratorPlugin instance.
func NewCIGeneratorPlugin() sdk.PluginProvider {
	return &ciGeneratorPlugin{}
}

// Manifest returns plugin metadata.
func (p *ciGeneratorPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "workflow-plugin-ci-generator",
		Version:     Version,
		Author:      "GoCodeAlone",
		Description: "CI/CD config generator for GitHub Actions, GitLab CI, Jenkins, and CircleCI",
	}
}

// TypedStepTypes returns the typed step type names this plugin provides.
func (p *ciGeneratorPlugin) TypedStepTypes() []string {
	return []string{
		"step.ci_generate",
	}
}

// CreateTypedStep creates a typed step instance of the given type.
func (p *ciGeneratorPlugin) CreateTypedStep(typeName, name string, config *anypb.Any) (sdk.StepInstance, error) {
	switch typeName {
	case "step.ci_generate":
		factory := sdk.NewTypedStepFactory(
			"step.ci_generate",
			&contracts.CIGenerateConfig{},
			&contracts.CIGenerateInput{},
			ExecuteCIGenerate,
		)
		return factory.CreateTypedStep(typeName, name, config)
	default:
		return nil, fmt.Errorf("ci-generator plugin: unknown step type %q", typeName)
	}
}

// ContractRegistry returns strict protobuf descriptors for plugin boundaries.
func (p *ciGeneratorPlugin) ContractRegistry() *pb.ContractRegistry {
	return &pb.ContractRegistry{
		FileDescriptorSet: &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				protodesc.ToFileDescriptorProto(contracts.File_internal_contracts_ci_generator_proto),
			},
		},
		Contracts: []*pb.ContractDescriptor{
			{
				Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
				StepType:      "step.ci_generate",
				ConfigMessage: "workflow.plugins.ci_generator.v1.CIGenerateConfig",
				InputMessage:  "workflow.plugins.ci_generator.v1.CIGenerateInput",
				OutputMessage: "workflow.plugins.ci_generator.v1.CIGenerateOutput",
				Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
			},
		},
	}
}
