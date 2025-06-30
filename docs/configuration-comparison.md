# Configuration Format Comparison

This document provides a side-by-side comparison of the legacy `bulk-clone.yaml` format and the new `gzh.yaml` format.

## Quick Comparison Table

| Feature | bulk-clone.yaml | gzh.yaml | Notes |
|---------|----------------|----------|-------|
| **Schema Version** | `version: "0.1"` | `version: "1.0.0"` | Updated version numbering |
| **Multi-Provider Support** | GitHub only | GitHub, GitLab, Gitea, Gogs | Unified configuration |
| **Authentication** | Protocol-based (`ssh`/`https`) | Token-based (`${GITHUB_TOKEN}`) | More secure and flexible |
| **Organization Structure** | `repo_roots[]` array | `providers.{provider}.orgs[]` | Better organization |
| **Global Excludes** | `ignore_names[]` | `exclude[]` per org | More granular control |
| **Directory Paths** | `root_path` | `clone_dir` | Clearer naming |
| **Filtering** | Basic ignore patterns | Regex `match` + `exclude` | Advanced filtering |
| **Clone Strategies** | Not supported | `reset`/`pull`/`fetch` | Flexible sync options |
| **Directory Structure** | Fixed nested | `flatten` option | Customizable layout |
| **Validation** | Limited | Built-in schema validation | Better error messages |

## Detailed Format Comparison

### Basic Configuration

#### bulk-clone.yaml
```yaml
version: "0.1"

default:
  protocol: https
  github:
    root_path: "$HOME/repos/github"

repo_roots:
  - root_path: "$HOME/work/mycompany"
    provider: "github"
    protocol: "ssh"
    org_name: "mycompany"

ignore_names:
  - "test-.*"
  - ".*-archive"
```

#### gzh.yaml
```yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "mycompany"
        clone_dir: "$HOME/work/mycompany"
        exclude: ["test-.*", ".*-archive"]
        strategy: "reset"
```

### Multi-Organization Setup

#### bulk-clone.yaml
```yaml
version: "0.1"

repo_roots:
  - root_path: "$HOME/work/frontend"
    provider: "github"
    protocol: "ssh"
    org_name: "frontend-team"
  
  - root_path: "$HOME/work/backend"
    provider: "github"
    protocol: "ssh"
    org_name: "backend-team"
  
  - root_path: "$HOME/opensource"
    provider: "github"
    protocol: "https"
    org_name: "kubernetes"

ignore_names:
  - "test-.*"
  - ".*-archive"
  - "temp-.*"
```

#### gzh.yaml
```yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "frontend-team"
        visibility: "private"
        clone_dir: "$HOME/work/frontend"
        match: "^(web|app|ui)-.*"
        exclude: ["test-.*", ".*-archive", "temp-.*"]
        strategy: "pull"
        flatten: true
      
      - name: "backend-team"
        visibility: "private"
        clone_dir: "$HOME/work/backend"
        match: "^(api|service|worker)-.*"
        exclude: ["test-.*", ".*-archive", "temp-.*"]
        strategy: "pull"
        flatten: true
      
      - name: "kubernetes"
        visibility: "public"
        clone_dir: "$HOME/opensource"
        exclude: ["test-.*", ".*-archive", "temp-.*"]
        strategy: "fetch"
```

## Field Mapping Reference

### Root Level Fields

| bulk-clone.yaml | gzh.yaml | Type | Notes |
|----------------|----------|------|-------|
| `version` | `version` | string | Version format changed |
| `default` | _(distributed)_ | object | Settings moved to specific contexts |
| `repo_roots` | `providers.{provider}.orgs` | array | Restructured by provider |
| `ignore_names` | _(per-org exclude)_ | array | Moved to organization level |

### Organization Level Fields

| bulk-clone.yaml | gzh.yaml | Type | Notes |
|----------------|----------|------|-------|
| `root_path` | `clone_dir` | string | Renamed for clarity |
| `provider` | _(provider section)_ | string | Now implicit from section |
| `protocol` | _(removed)_ | string | Replaced by token authentication |
| `org_name` | `name` | string | Simplified naming |
| _(not supported)_ | `visibility` | string | New: `public`/`private`/`all` |
| _(not supported)_ | `match` | string | New: Regex filtering |
| _(not supported)_ | `exclude` | array | New: Per-org exclusions |
| _(not supported)_ | `strategy` | string | New: `reset`/`pull`/`fetch` |
| _(not supported)_ | `flatten` | boolean | New: Directory structure control |
| _(not supported)_ | `recursive` | boolean | New: GitLab subgroup support |

### Default Settings Migration

| bulk-clone.yaml | gzh.yaml | Notes |
|----------------|----------|-------|
| `default.protocol` | _(removed)_ | Authentication via tokens |
| `default.github.root_path` | _(per-org clone_dir)_ | More flexible per-organization |
| `default.gitlab.url` | _(per-provider)_ | Provider-specific settings |
| `default.gitlab.recursive` | _(per-group)_ | More granular control |

## Feature Comparison

### Authentication

| Feature | bulk-clone.yaml | gzh.yaml |
|---------|----------------|----------|
| **Token Support** | ❌ No | ✅ Yes |
| **SSH Keys** | ✅ Via protocol | ❌ Use tokens instead |
| **HTTPS** | ✅ Via protocol | ✅ Via tokens |
| **Environment Variables** | ❌ No | ✅ Yes |
| **Per-Org Tokens** | ❌ No | ✅ Yes |

### Repository Filtering

| Feature | bulk-clone.yaml | gzh.yaml |
|---------|----------------|----------|
| **Global Excludes** | ✅ `ignore_names` | ❌ Use per-org excludes |
| **Per-Org Excludes** | ❌ No | ✅ `exclude` |
| **Regex Matching** | ✅ Basic | ✅ Full regex support |
| **Include Patterns** | ❌ No | ✅ `match` field |
| **Visibility Filter** | ❌ No | ✅ `public`/`private`/`all` |

### Provider Support

| Provider | bulk-clone.yaml | gzh.yaml |
|----------|----------------|----------|
| **GitHub** | ✅ Full support | ✅ Enhanced support |
| **GitLab** | ⚠️ Planned | ✅ Full support |
| **Gitea** | ❌ No | ✅ Full support |
| **Gogs** | ❌ No | ✅ Full support |
| **Multi-Provider** | ❌ No | ✅ Yes |

### Directory Management

| Feature | bulk-clone.yaml | gzh.yaml |
|---------|----------------|----------|
| **Custom Paths** | ✅ `root_path` | ✅ `clone_dir` |
| **Environment Variables** | ✅ Yes | ✅ Yes |
| **Nested Structure** | ✅ Default | ✅ `flatten: false` |
| **Flat Structure** | ❌ No | ✅ `flatten: true` |
| **Per-Org Paths** | ✅ Yes | ✅ Yes |

### Update Strategies

| Strategy | bulk-clone.yaml | gzh.yaml |
|----------|----------------|----------|
| **Clone Only** | ✅ Default | ✅ First run |
| **Hard Reset** | ❌ No | ✅ `strategy: reset` |
| **Git Pull** | ❌ No | ✅ `strategy: pull` |
| **Git Fetch** | ❌ No | ✅ `strategy: fetch` |
| **Custom Strategy** | ❌ No | ⚠️ Planned |

## Migration Complexity

### Simple Migration (Low Complexity)
- Single GitHub organization
- Basic path configuration
- Global ignore patterns
- No advanced features

**Effort**: 5-10 minutes

### Medium Migration (Medium Complexity)
- Multiple GitHub organizations
- Different paths per organization
- Mix of public/private repositories
- Some filtering requirements

**Effort**: 15-30 minutes

### Complex Migration (High Complexity)
- Multiple providers (GitHub + GitLab)
- Advanced filtering requirements
- Custom authentication setups
- Complex directory structures

**Effort**: 30-60 minutes + testing

## Benefits of Migration

### Immediate Benefits
- ✅ **Better Security**: Token-based authentication
- ✅ **Multi-Provider**: Support for GitLab, Gitea, Gogs
- ✅ **Better Validation**: Schema validation with clear error messages
- ✅ **Granular Control**: Per-organization settings

### Advanced Benefits
- ✅ **Flexible Filtering**: Regex matching and visibility filters
- ✅ **Directory Control**: Flatten option for better organization
- ✅ **Update Strategies**: Choose how repositories are updated
- ✅ **Environment Integration**: Better environment variable support

### Future Benefits
- 🔄 **Performance**: Optimized for large-scale operations
- 🔄 **Extensions**: Plugin system for custom providers
- 🔄 **Automation**: Better CI/CD integration
- 🔄 **Monitoring**: Built-in progress tracking and reporting

## Migration Tools and Support

### Available Tools
- 📜 **Migration Script**: `scripts/migrate-config.sh`
- 📖 **Migration Guide**: `docs/migration-guide-bulk-clone-to-gzh.md`
- 🔍 **Validation Tool**: `gzh config validate`
- 🧪 **Dry Run**: `gzh bulk-clone --dry-run`

### Getting Help
- 📚 **Documentation**: Comprehensive guides and examples
- 🐛 **Issue Tracker**: Report migration problems
- 💬 **Community**: Discussion forums and support channels
- 🔧 **Professional Support**: Enterprise migration assistance

---

For step-by-step migration instructions, see the [Migration Guide](./migration-guide-bulk-clone-to-gzh.md).