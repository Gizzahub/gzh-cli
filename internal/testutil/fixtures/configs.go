package fixtures

// SampleBulkCloneConfig provides a sample bulk-clone configuration for testing
const SampleBulkCloneConfig = `version: "1.0"
default_provider: github
providers:
  github:
    token: ${GITHUB_TOKEN}
    orgs:
      - name: test-org
        visibility: public
        clone_dir: ./repos/github/test-org
        exclude:
          - private-repo
        strategy: reset
  gitlab:
    token: ${GITLAB_TOKEN}
    groups:
      - name: test-group
        recursive: true
        clone_dir: ./repos/gitlab/test-group
        strategy: pull
`

// MinimalConfig provides a minimal valid configuration
const MinimalConfig = `version: "1.0"
providers:
  github:
    token: test-token
    orgs:
      - name: test-org
`

// InvalidConfig provides an invalid configuration for error testing
const InvalidConfig = `version: ""
providers:
  github:
    orgs:
      - name: test-org
`

// ComplexConfig provides a complex configuration with all features
const ComplexConfig = `version: "1.0"
default_provider: github
providers:
  github:
    token: ${GITHUB_TOKEN}
    orgs:
      - name: org1
        visibility: all
        clone_dir: ./repos/github/org1
        match: "^(app-|lib-)"
        exclude:
          - deprecated-repo
          - archive-repo
        strategy: reset
      - name: org2
        visibility: public
        clone_dir: ./repos/github/org2
        strategy: fetch
  gitlab:
    token: ${GITLAB_TOKEN}
    groups:
      - name: group1
        recursive: true
        flatten: true
        clone_dir: ./repos/gitlab/group1
        strategy: pull
  gitea:
    token: ${GITEA_TOKEN}
    orgs:
      - name: gitea-org
        clone_dir: ./repos/gitea/gitea-org
`

// RepoConfigSample provides a sample repository configuration
const RepoConfigSample = `name: test-repo
description: Test repository configuration
visibility: public
topics:
  - golang
  - testing
settings:
  has_issues: true
  has_wiki: false
  has_projects: false
branch_protection:
  - pattern: main
    required_reviews: 2
    dismiss_stale_reviews: true
    require_code_owner_reviews: true
`
