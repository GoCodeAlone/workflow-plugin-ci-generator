# workflow-plugin-ci-generator

> ⚠️ **Experimental** — This plugin compiles and passes its unit tests but has not been validated in any active GoCodeAlone-internal production deployment. Use with caution. Please [open an issue](https://github.com/GoCodeAlone/workflow-plugin-ci-generator/issues/new) if you adopt it so we can promote it to **verified** status.

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/GoCodeAlone/workflow-plugin-ci-generator.svg)](https://pkg.go.dev/github.com/GoCodeAlone/workflow-plugin-ci-generator)

CI/CD config generator for workflow projects — emits GitHub Actions, GitLab CI, Jenkins, and CircleCI pipelines from workflow project manifests.

## What it provides

**Pipeline step types:**
- `step.ci_generate` — Generate CI/CD configuration files (GitHub Actions, GitLab CI, Jenkins, CircleCI) from a workflow project manifest

## Install

```yaml
# In your wfctl.yaml
version: 1
plugins:
  - name: workflow-plugin-ci-generator
    version: v0.1.3
    source: github.com/GoCodeAlone/workflow-plugin-ci-generator
```

Then:

```sh
wfctl plugin install
```

## Minimal example

See [`examples/minimal/config.yaml`](examples/minimal/config.yaml).

## Supported CI platforms

| Platform | Output file |
|----------|-------------|
| GitHub Actions | `.github/workflows/workflow.yml` |
| GitLab CI | `.gitlab-ci.yml` |
| Jenkins | `Jenkinsfile` |
| CircleCI | `.circleci/config.yml` |

## Documentation

- [Plugin authoring guide (upstream)](https://github.com/GoCodeAlone/workflow/blob/main/docs/PLUGIN_AUTHORING.md)
- [Workflow engine docs](https://github.com/GoCodeAlone/workflow)

## License

MIT. See [LICENSE](LICENSE).
