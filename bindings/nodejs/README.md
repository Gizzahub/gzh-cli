# GZH Manager Node.js Bindings

Node.js bindings for GZH Manager - A comprehensive CLI tool for managing development environments and Git repositories across multiple platforms.

## Features

- 🚀 **Bulk Repository Cloning** - Clone entire organizations from GitHub, GitLab, Gitea, and Gogs
- 🔌 **Plugin System** - Extensible architecture with custom plugin support
- 🌐 **Multi-Platform Support** - Works with multiple Git hosting platforms
- 🛡️ **Security First** - Built-in security scanning and best practices
- 📊 **Progress Tracking** - Real-time progress updates for long-running operations
- 🎯 **Advanced Filtering** - Filter repositories by stars, size, language, and more
- 💪 **TypeScript Support** - Full TypeScript definitions included

## Installation

```bash
npm install @gizzahub/gzh-manager
```

### Prerequisites

- Node.js 16.0.0 or higher
- GZH Manager Go library (automatically built during installation)
- Platform tokens for accessing private repositories

## Quick Start

```typescript
import { createClient, PlatformConfig, BulkCloneRequest } from '@gizzahub/gzh-manager';

async function main() {
  // Create a client
  const client = createClient({
    timeout: 60,
    enablePlugins: true,
    logLevel: 'info',
  });

  try {
    // Configure platforms
    const platforms: PlatformConfig[] = [
      {
        type: 'github',
        token: process.env.GITHUB_TOKEN,
        organizations: ['octocat', 'github'],
        skipArchived: true,
      },
    ];

    // Perform bulk clone
    const result = await client.bulkClone({
      platforms,
      outputDir: './cloned-repos',
      concurrency: 5,
      strategy: 'reset',
      filters: {
        minStars: 10,
        languages: ['TypeScript', 'JavaScript'],
      },
    });

    if (result.success) {
      console.log(`Cloned ${result.data.cloned} repositories!`);
    }
  } finally {
    client.destroy();
  }
}

main().catch(console.error);
```

## API Reference

### Client

#### Constructor

```typescript
new Client(config?: ClientConfig)
```

Create a new GZH Manager client with optional configuration.

#### Methods

##### `bulkClone(request: BulkCloneRequest): Promise<OperationResult<CloneResult>>`

Perform bulk clone operation across multiple platforms.

```typescript
const result = await client.bulkClone({
  platforms: [
    {
      type: 'github',
      token: 'ghp_xxx',
      organizations: ['microsoft', 'google'],
    },
  ],
  outputDir: './repos',
  concurrency: 10,
  strategy: 'reset',
  includePrivate: false,
  filters: {
    minStars: 100,
    maxSize: 500,
    languages: ['Go', 'TypeScript'],
  },
});
```

##### `listPlugins(): Promise<OperationResult<PluginInfo[]>>`

List available plugins in the system.

```typescript
const plugins = await client.listPlugins();
if (plugins.success) {
  plugins.data.forEach(plugin => {
    console.log(`${plugin.name}: ${plugin.description}`);
  });
}
```

##### `executePlugin(request: PluginExecuteRequest): Promise<OperationResult<any>>`

Execute a plugin method with arguments.

```typescript
const result = await client.executePlugin({
  pluginName: 'code-analyzer',
  method: 'analyze',
  args: { path: './project', includeTests: true },
  timeout: 30,
});
```

##### `health(): Promise<OperationResult<HealthInfo>>`

Get client health and system information.

```typescript
const health = await client.health();
if (health.success) {
  console.log(`Status: ${health.data.status}`);
  console.log(`Version: ${health.data.version}`);
}
```

##### `destroy(): void`

Clean up client resources. Always call this when done.

### Configuration Types

#### `ClientConfig`

```typescript
interface ClientConfig {
  timeout?: number;          // Request timeout in seconds
  retryCount?: number;       // Number of retry attempts
  enablePlugins?: boolean;   // Enable plugin system
  pluginDir?: string;        // Custom plugin directory
  logLevel?: string;         // Log level (debug, info, warn, error)
  logFile?: string;          // Log file path
}
```

#### `PlatformConfig`

```typescript
interface PlatformConfig {
  type: 'github' | 'gitlab' | 'gitea' | 'gogs';
  baseUrl?: string;          // Custom base URL for self-hosted instances
  token?: string;            // Authentication token
  organizations: string[];   // Organizations/groups to clone
  skipArchived?: boolean;    // Skip archived repositories
  skipForked?: boolean;      // Skip forked repositories
}
```

#### `CloneFilters`

```typescript
interface CloneFilters {
  include?: string[];        // Include patterns
  exclude?: string[];        // Exclude patterns
  minStars?: number;         // Minimum star count
  maxSize?: number;          // Maximum size in MB
  languages?: string[];      // Filter by programming languages
  updatedAfter?: string;     // Only repos updated after date
}
```

## Examples

### Multi-Platform Cloning

```typescript
import { Client } from '@gizzahub/gzh-manager';

const client = new Client();

const platforms = [
  {
    type: 'github',
    token: process.env.GITHUB_TOKEN,
    organizations: ['microsoft', 'google'],
  },
  {
    type: 'gitlab',
    baseUrl: 'https://gitlab.com',
    token: process.env.GITLAB_TOKEN,
    organizations: ['gitlab-org'],
  },
];

const result = await client.bulkClone({
  platforms,
  outputDir: './multi-platform-repos',
  concurrency: 8,
});
```

### Plugin Usage

```typescript
// List available plugins
const plugins = await client.listPlugins();

// Execute a security scan plugin
const scanResult = await client.executePlugin({
  pluginName: 'security-scanner',
  method: 'scan_vulnerabilities',
  args: {
    path: './project',
    severity: 'high',
  },
});
```

### Error Handling

```typescript
try {
  const result = await client.bulkClone(request);
  
  if (!result.success) {
    console.error('Clone failed:', result.error);
    return;
  }
  
  // Handle successful result
  const { cloned, failed, errors } = result.data;
  console.log(`Cloned: ${cloned}, Failed: ${failed}`);
  
  // Log any errors
  errors.forEach(error => {
    console.error(`${error.repository}: ${error.error}`);
  });
  
} catch (error) {
  console.error('Unexpected error:', error);
} finally {
  client.destroy();
}
```

## Environment Variables

Set these environment variables for platform authentication:

```bash
export GITHUB_TOKEN="ghp_your_github_token"
export GITLAB_TOKEN="glpat_your_gitlab_token"
export GITEA_TOKEN="your_gitea_token"
export GOGS_TOKEN="your_gogs_token"
```

## Building from Source

```bash
# Clone the repository
git clone https://github.com/gizzahub/gzh-manager-go.git
cd gzh-manager-go/bindings/nodejs

# Install dependencies
npm install

# Build the project
npm run build

# Run tests
npm test
```

## Platform Support

| Platform | Support | Features |
|----------|---------|----------|
| GitHub | ✅ Full | Organizations, repos, releases |
| GitLab | ✅ Full | Groups, projects, CI/CD |
| Gitea | ✅ Full | Organizations, repositories |
| Gogs | 🚧 Planned | Basic repository cloning |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://github.com/gizzahub/gzh-manager-go/wiki)
- 🐛 [Issue Tracker](https://github.com/gizzahub/gzh-manager-go/issues)
- 💬 [Discussions](https://github.com/gizzahub/gzh-manager-go/discussions)

## Related Projects

- [GZH Manager CLI](https://github.com/gizzahub/gzh-manager-go) - The main Go CLI tool
- [GZH Manager Python](../python/) - Python bindings
- [GZH Manager Plugins](https://github.com/gizzahub/gzh-plugins) - Official plugin collection