# GitHub Organization and Repository Management Research

## Overview

Research notes for implementing GitHub organization and repository configuration management functionality. This could potentially be implemented using either Terraform or GitHub Actions, though GitHub Actions may not be the best fit for this use case.

## Reference Projects

### GitHub Actions Examples

- [Hello World Docker Action](https://github.com/actions/hello-world-docker-action)
- [TypeScript Action](https://github.com/actions/typescript-action)
- [Hello World JavaScript Action](https://github.com/actions/hello-world-javascript-action)
- [JavaScript Action](https://github.com/actions/javascript-action)
- [Starter Workflows](https://github.com/actions/starter-workflows)

## Reference Documentation

### GitHub Actions Security

- [Automatic Token Authentication](https://docs.github.com/ko/actions/security-for-github-actions/security-guides/automatic-token-authentication)
- [Repository API Documentation](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#update-a-repository)

### Environment Variables in GitHub Actions

- [Using env variables as default values](https://stackoverflow.com/questions/73955908/how-to-use-env-variable-as-default-value-for-input-in-github-actions)

### Token Permissions

- [Permissions Required for GitHub Apps](https://docs.github.com/ko/rest/authentication/permissions-required-for-github-apps?apiVersion=2022-11-28)
- [Automatic Token Authentication](https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication)

## Implementation Reference

### Repository Update API Example (Octokit)

```javascript
const repoUpdateResult = octokit.repos.update({
  owner: repoOwner,
  repo: repo.name,

  name: repo.name,
  // description: repo.description,
  // homepage: repo.homepage,
  private: repoMeta.private,
  visibility: repoMeta.visibility,
  security_and_analysis: repoMeta.security_and_analysis,

  has_issues: repoMeta.has_issues,
  has_projects: repoMeta.has_projects,
  has_wiki: repoMeta.has_wiki,

  default_branch: repo.default_branch,

  allow_squash_merge: repoMeta.allow_squash_merge,
  allow_merge_commit: repoMeta.allow_merge_commit,
  allow_rebase_merge: repoMeta.allow_rebase_merge,

  delete_branch_on_merge: repoMeta.delete_branch_on_merge,

  allow_update_branch: repoMeta.allow_update_branch,

  use_squash_pr_title_as_default: repoMeta.use_squash_pr_title_as_default,

  squash_merge_commit_title: repoMeta.squash_merge_commit_title,
  squash_merge_commit_message: repoMeta.squash_merge_commit_message,

  merge_commit_title: repoMeta.merge_commit_title,
  merge_commit_message: repoMeta.merge_commit_message,

  archived: repoMeta.archived,
  allow_forking: repoMeta.allow_forking,
  allow_auto_merge: repoMeta.allow_auto_merge,

  web_commit_signoff_required: repoMeta.web_commit_signoff_required,
});
```

## Implementation Notes

This functionality could be valuable for:

- Bulk repository configuration management
- Organization-wide policy enforcement
- Standardizing repository settings across multiple repos
- Automating repository setup for new projects

Consider implementing as part of the existing `gz` CLI rather than separate GitHub Actions.
