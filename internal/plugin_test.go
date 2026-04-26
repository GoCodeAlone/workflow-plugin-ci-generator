package internal

import (
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-ci-generator/internal/contracts"
	pb "github.com/GoCodeAlone/workflow/plugin/external/proto"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestNewCIGeneratorPluginTypedContracts(t *testing.T) {
	provider := NewCIGeneratorPlugin()
	if _, ok := provider.(sdk.TypedStepProvider); !ok {
		t.Fatal("expected typed step provider")
	}
	contractProvider, ok := provider.(sdk.ContractProvider)
	if !ok {
		t.Fatal("expected contract provider")
	}

	registry := contractProvider.ContractRegistry()
	if registry == nil || len(registry.Contracts) != 1 {
		t.Fatalf("expected one contract descriptor, got %#v", registry)
	}
	descriptor := registry.Contracts[0]
	if descriptor.Kind != pb.ContractKind_CONTRACT_KIND_STEP {
		t.Fatalf("unexpected kind: %s", descriptor.Kind)
	}
	if descriptor.StepType != "step.ci_generate" {
		t.Fatalf("unexpected step type: %s", descriptor.StepType)
	}
	if descriptor.Mode != pb.ContractMode_CONTRACT_MODE_STRICT_PROTO {
		t.Fatalf("unexpected contract mode: %s", descriptor.Mode)
	}
	if registry.FileDescriptorSet == nil || len(registry.FileDescriptorSet.File) == 0 {
		t.Fatal("expected file descriptor set for plugin-owned messages")
	}
}

func TestCreateTypedStepValidatesConfig(t *testing.T) {
	provider := NewCIGeneratorPlugin().(sdk.TypedStepProvider)
	config, err := anypb.New(&contracts.CIGenerateConfig{Platform: PlatformGitHubActions})
	if err != nil {
		t.Fatalf("pack config: %v", err)
	}

	step, err := provider.CreateTypedStep("step.ci_generate", "generate", config)
	if err != nil {
		t.Fatalf("CreateTypedStep: %v", err)
	}
	if step == nil {
		t.Fatal("expected typed step instance")
	}
}
