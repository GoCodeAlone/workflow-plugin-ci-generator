package platforms

import (
	"bytes"
	"fmt"
	"text/template"
)

// CircleCIGenerator generates a .circleci/config.yml file using CircleCI v2.1 syntax.
type CircleCIGenerator struct{}

// NewCircleCIGenerator returns a new CircleCIGenerator.
func NewCircleCIGenerator() *CircleCIGenerator {
	return &CircleCIGenerator{}
}

// Generate produces:
//   - .circleci/config.yml
func (g *CircleCIGenerator) Generate(opts Options) (map[string]string, error) {
	branch := opts.DefaultBranch
	if branch == "" {
		branch = "main"
	}
	infra := opts.InfraConfig
	if infra == "" {
		infra = "infra.yaml"
	}

	data := circleciData{
		DefaultBranch: branch,
		InfraConfig:   infra,
		ProjectName:   opts.ProjectName,
	}

	content, err := renderCircleCITemplate(circleciConfigTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("circleci: render config.yml: %w", err)
	}

	return map[string]string{
		".circleci/config.yml": content,
	}, nil
}

type circleciData struct {
	DefaultBranch string
	InfraConfig   string
	ProjectName   string
}

func renderCircleCITemplate(tmplStr string, data circleciData) (string, error) {
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

// circleciConfigTemplate uses CircleCI v2.1 syntax with orbs, executors, and
// an approval job for the apply step.
const circleciConfigTemplate = `version: 2.1

orbs:
  go: circleci/go@1.11

executors:
  default:
    docker:
      - image: cimg/go:1.26
    resource_class: medium

jobs:
  infra-plan:
    executor: default
    steps:
      - checkout
      - run:
          name: Install wfctl
          command: curl -fsSL https://github.com/GoCodeAlone/workflow/releases/latest/download/wfctl-linux-amd64.tar.gz | tar -xz && sudo mv wfctl /usr/local/bin/
      - run:
          name: Plan infrastructure
          command: |
            wfctl infra plan -c {{.InfraConfig}} --output plan.json
            wfctl infra plan -c {{.InfraConfig}} --format markdown > plan.md
      - store_artifacts:
          path: plan.json
      - store_artifacts:
          path: plan.md
      - persist_to_workspace:
          root: .
          paths:
            - plan.json

  approve-apply:
    executor: default
    steps:
      - run: echo "Waiting for manual approval before applying infrastructure changes."

  infra-apply:
    executor: default
    steps:
      - checkout
      - attach_workspace:
          at: .
      - run:
          name: Install wfctl
          command: curl -fsSL https://github.com/GoCodeAlone/workflow/releases/latest/download/wfctl-linux-amd64.tar.gz | tar -xz && sudo mv wfctl /usr/local/bin/
      - run:
          name: Apply infrastructure
          command: wfctl infra apply -c {{.InfraConfig}} --auto-approve

  build:
    executor: default
    steps:
      - checkout
      - go/load-cache
      - go/mod-download
      - go/save-cache
      - run:
          name: Test
          command: go test ./...
      - run:
          name: Build
          command: go build ./...
      - run:
          name: Build and push container image
          command: |
            docker build -t $REGISTRY_IMAGE:$CIRCLE_SHA1 .
            docker push $REGISTRY_IMAGE:$CIRCLE_SHA1

  deploy:
    executor: default
    steps:
      - checkout
      - run:
          name: Install wfctl
          command: curl -fsSL https://github.com/GoCodeAlone/workflow/releases/latest/download/wfctl-linux-amd64.tar.gz | tar -xz && sudo mv wfctl /usr/local/bin/
      - run:
          name: Deploy application
          command: wfctl deploy --image $REGISTRY_IMAGE:$CIRCLE_SHA1

workflows:
  infra-and-deploy:
    jobs:
      - infra-plan:
          filters:
            branches:
              only:
                - {{.DefaultBranch}}
      - approve-apply:
          type: approval
          requires:
            - infra-plan
          filters:
            branches:
              only:
                - {{.DefaultBranch}}
      - infra-apply:
          requires:
            - approve-apply
          filters:
            branches:
              only:
                - {{.DefaultBranch}}
      - build:
          filters:
            branches:
              only:
                - {{.DefaultBranch}}
      - deploy:
          requires:
            - infra-apply
            - build
          filters:
            branches:
              only:
                - {{.DefaultBranch}}
`
