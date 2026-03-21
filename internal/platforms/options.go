// Package platforms provides CI/CD config generators for each supported platform.
package platforms

// Options holds the common inputs all platform generators use.
type Options struct {
	// InfraConfig is the path to the infra.yaml file (e.g. "infra.yaml").
	InfraConfig string
	// ProjectName is used as a label in generated configs.
	ProjectName string
	// Runner is the CI runner label (used by GitHub Actions; ignored elsewhere).
	Runner string
	// DefaultBranch is the main branch name (default: "main").
	DefaultBranch string
}
