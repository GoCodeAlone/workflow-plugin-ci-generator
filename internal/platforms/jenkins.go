package platforms

import (
	"bytes"
	"fmt"
	"text/template"
)

// JenkinsGenerator generates a declarative Jenkinsfile.
type JenkinsGenerator struct{}

// NewJenkinsGenerator returns a new JenkinsGenerator.
func NewJenkinsGenerator() *JenkinsGenerator {
	return &JenkinsGenerator{}
}

// Generate produces:
//   - Jenkinsfile
func (g *JenkinsGenerator) Generate(opts Options) (map[string]string, error) {
	branch := opts.DefaultBranch
	if branch == "" {
		branch = "main"
	}
	infra := opts.InfraConfig
	if infra == "" {
		infra = "infra.yaml"
	}

	data := jenkinsData{
		DefaultBranch: branch,
		InfraConfig:   infra,
		ProjectName:   opts.ProjectName,
	}

	content, err := renderJenkinsTemplate(jenkinsfileTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("jenkins: render Jenkinsfile: %w", err)
	}

	return map[string]string{
		"Jenkinsfile": content,
	}, nil
}

type jenkinsData struct {
	DefaultBranch string
	InfraConfig   string
	ProjectName   string
}

func renderJenkinsTemplate(tmplStr string, data jenkinsData) (string, error) {
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

// jenkinsfileTemplate uses declarative pipeline syntax (not scripted).
const jenkinsfileTemplate = `pipeline {
    agent {
        label 'linux'
    }

    environment {
        INFRA_CONFIG = '{{.InfraConfig}}'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Plan') {
            when {
                changeRequest()
            }
            steps {
                sh 'wfctl infra plan -c $INFRA_CONFIG --output plan.json'
                sh 'wfctl infra plan -c $INFRA_CONFIG --format markdown > plan.md'
            }
            post {
                always {
                    archiveArtifacts artifacts: 'plan.json,plan.md', allowEmptyArchive: true
                }
            }
        }

        stage('Build') {
            steps {
                sh 'go test ./...'
                sh 'go build ./...'
                sh 'docker build -t $REGISTRY_IMAGE:$GIT_COMMIT .'
            }
        }

        stage('Push Image') {
            when {
                branch '{{.DefaultBranch}}'
            }
            steps {
                sh 'docker push $REGISTRY_IMAGE:$GIT_COMMIT'
            }
        }

        stage('Apply') {
            when {
                branch '{{.DefaultBranch}}'
            }
            steps {
                sh 'wfctl infra apply -c $INFRA_CONFIG --auto-approve'
            }
        }

        stage('Deploy') {
            when {
                branch '{{.DefaultBranch}}'
            }
            steps {
                sh 'wfctl deploy --image $REGISTRY_IMAGE:$GIT_COMMIT'
            }
        }
    }

    post {
        failure {
            echo 'Pipeline failed. Check logs above.'
        }
        success {
            echo 'Pipeline completed successfully.'
        }
    }
}
`
