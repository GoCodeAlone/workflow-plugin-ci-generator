package platforms

import (
	"bytes"
	"fmt"
	"text/template"
)

// GitLabCIGenerator generates a .gitlab-ci.yml file using GitLab CI v17+ syntax.
type GitLabCIGenerator struct{}

// NewGitLabCIGenerator returns a new GitLabCIGenerator.
func NewGitLabCIGenerator() *GitLabCIGenerator {
	return &GitLabCIGenerator{}
}

// Generate produces:
//   - .gitlab-ci.yml
func (g *GitLabCIGenerator) Generate(opts Options) (map[string]string, error) {
	branch := opts.DefaultBranch
	if branch == "" {
		branch = "main"
	}
	infra := opts.InfraConfig
	if infra == "" {
		infra = "infra.yaml"
	}

	data := gitlabData{
		DefaultBranch: branch,
		InfraConfig:   infra,
		ProjectName:   opts.ProjectName,
	}

	content, err := renderGitLabTemplate(gitlabCITemplate, data)
	if err != nil {
		return nil, fmt.Errorf("gitlab ci: render .gitlab-ci.yml: %w", err)
	}

	return map[string]string{
		".gitlab-ci.yml": content,
	}, nil
}

type gitlabData struct {
	DefaultBranch string
	InfraConfig   string
	ProjectName   string
}

func renderGitLabTemplate(tmplStr string, data gitlabData) (string, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// gitlabCITemplate uses GitLab CI v17+ syntax:
//   - rules: (not deprecated only:)
//   - needs: for DAG pipeline ordering
//   - environment: for deployment tracking
const gitlabCITemplate = `stages:
  - plan
  - apply
  - build
  - deploy

default:
  image: golang:1.26

variables:
  INFRA_CONFIG: "{{.InfraConfig}}"

# ── Plan ────────────────────────────────────────────────────────────────────
infra-plan:
  stage: plan
  script:
    - wfctl infra plan -c "$INFRA_CONFIG" --output plan.json
    - wfctl infra plan -c "$INFRA_CONFIG" --format markdown > plan.md
  artifacts:
    paths:
      - plan.json
      - plan.md
    expire_in: 1 hour
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
      changes:
        - "{{.InfraConfig}}"
        - "infra/**/*"

# ── Apply ───────────────────────────────────────────────────────────────────
infra-apply:
  stage: apply
  needs:
    - job: infra-plan
      artifacts: true
  script:
    - wfctl infra apply -c "$INFRA_CONFIG" --auto-approve
  environment:
    name: production
    url: https://your-app.example.com
  rules:
    - if: $CI_COMMIT_BRANCH == "{{.DefaultBranch}}" && $CI_PIPELINE_SOURCE == "push"
      changes:
        - "{{.InfraConfig}}"
        - "infra/**/*"

# ── Build ───────────────────────────────────────────────────────────────────
build:
  stage: build
  needs: []
  script:
    - go test ./...
    - go build ./...
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  rules:
    - if: $CI_COMMIT_BRANCH == "{{.DefaultBranch}}"

# ── Deploy ──────────────────────────────────────────────────────────────────
deploy:
  stage: deploy
  needs:
    - job: infra-apply
    - job: build
  script:
    - wfctl deploy --image $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  environment:
    name: production
    url: https://your-app.example.com
  rules:
    - if: $CI_COMMIT_BRANCH == "{{.DefaultBranch}}"
`
