package platforms

import (
	"bytes"
	"fmt"
	"text/template"
)

// GitHubActionsGenerator generates GitHub Actions workflow files.
type GitHubActionsGenerator struct{}

// NewGitHubActionsGenerator returns a new GitHubActionsGenerator.
func NewGitHubActionsGenerator() *GitHubActionsGenerator {
	return &GitHubActionsGenerator{}
}

// Generate produces the following files:
//   - .github/workflows/infra.yml  — plan on PR, apply on push to main
//   - .github/workflows/build.yml  — build, test, push container
//   - .github/workflows/deploy.yml — deploy after infra apply
func (g *GitHubActionsGenerator) Generate(opts Options) (map[string]string, error) {
	runner := opts.Runner
	if runner == "" {
		runner = "self-hosted, Linux, X64"
	}
	branch := opts.DefaultBranch
	if branch == "" {
		branch = "main"
	}
	infra := opts.InfraConfig
	if infra == "" {
		infra = "infra.yaml"
	}

	data := ghaData{
		Runner:        runner,
		DefaultBranch: branch,
		InfraConfig:   infra,
		ProjectName:   opts.ProjectName,
	}

	files := map[string]string{}

	infraYAML, err := renderGHATemplate(ghaInfraTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("github actions: render infra.yml: %w", err)
	}
	files[".github/workflows/infra.yml"] = infraYAML

	buildYAML, err := renderGHATemplate(ghaBuildTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("github actions: render build.yml: %w", err)
	}
	files[".github/workflows/build.yml"] = buildYAML

	deployYAML, err := renderGHATemplate(ghaDeployTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("github actions: render deploy.yml: %w", err)
	}
	files[".github/workflows/deploy.yml"] = deployYAML

	return files, nil
}

type ghaData struct {
	Runner        string
	DefaultBranch string
	InfraConfig   string
	ProjectName   string
}

func renderGHATemplate(tmplStr string, data ghaData) (string, error) {
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

// ghaInfraTemplate is the infra.yml template: plan on PR, apply on push to main.
const ghaInfraTemplate = `name: Infrastructure
on:
  pull_request:
    paths:
      - '{{.InfraConfig}}'
      - 'infra/**'
  push:
    branches:
      - {{.DefaultBranch}}
    paths:
      - '{{.InfraConfig}}'
      - 'infra/**'
permissions:
  contents: read
  pull-requests: write
jobs:
  plan:
    if: github.event_name == 'pull_request'
    runs-on: [{{.Runner}}]
    steps:
      - uses: actions/checkout@v4
      - uses: GoCodeAlone/setup-wfctl@v1
      - name: Plan infrastructure
        run: wfctl infra plan -c {{.InfraConfig}} --output plan.json
      - name: Format plan as markdown
        run: wfctl infra plan -c {{.InfraConfig}} --format markdown > plan.md
      - name: Post plan comment
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const plan = fs.readFileSync('plan.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## Infrastructure Plan\n\n' + plan
            });
  apply:
    if: github.event_name == 'push' && github.ref == 'refs/heads/{{.DefaultBranch}}'
    runs-on: [{{.Runner}}]
    steps:
      - uses: actions/checkout@v4
      - uses: GoCodeAlone/setup-wfctl@v1
      - name: Apply infrastructure
        run: wfctl infra apply -c {{.InfraConfig}} --auto-approve
`

// ghaBuildTemplate is the build.yml template: build, test, push container.
const ghaBuildTemplate = `name: Build
on:
  push:
    branches:
      - {{.DefaultBranch}}
  pull_request:
    branches:
      - {{.DefaultBranch}}
permissions:
  contents: read
  packages: write
jobs:
  build:
    runs-on: [{{.Runner}}]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
      - name: Run tests
        run: go test ./...
      - name: Build
        run: go build ./...
      - name: Log in to GitHub Container Registry
        if: github.event_name == 'push'
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{"{{"}} github.actor {{"}}"}}
          password: ${{"{{"}} secrets.GITHUB_TOKEN {{"}}"}}
      - name: Build and push container image
        if: github.event_name == 'push'
        uses: docker/build-push-action@v6
        with:
          push: true
          tags: ghcr.io/${{"{{"}} github.repository {{"}}"}}:${{"{{"}} github.sha {{"}}"}}
`

// ghaDeployTemplate is the deploy.yml template: deploy after infra apply.
const ghaDeployTemplate = `name: Deploy
on:
  workflow_run:
    workflows:
      - Infrastructure
    types:
      - completed
    branches:
      - {{.DefaultBranch}}
permissions:
  contents: read
jobs:
  deploy:
    if: github.event.workflow_run.conclusion == 'success'
    runs-on: [{{.Runner}}]
    steps:
      - uses: actions/checkout@v4
      - uses: GoCodeAlone/setup-wfctl@v1
      - name: Deploy application
        run: |
          wfctl infra apply -c {{.InfraConfig}} --auto-approve
          wfctl deploy --image ghcr.io/${{"{{"}} github.repository {{"}}"}}:${{"{{"}} github.sha {{"}}"}}
`
