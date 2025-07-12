# Security Policy

## Known Deprecation Warnings

This project currently shows deprecation warnings for the following reasons:

### React Scripts Dependencies
- **react-scripts 5.0.1** includes several outdated dependencies
- These are transitive dependencies not directly used by our code
- Upgrading to newer build tools (like Vite) would resolve these warnings

### Deprecated Packages Status

| Package | Status | Resolution |
|---------|--------|------------|
| `inflight` | Deprecated, memory leak | Indirect dependency via react-scripts |
| `stable` | Deprecated | Modern JS has stable sort |
| Babel proposal plugins | Merged to ES standard | Used by react-scripts build tools |
| `eslint` 8.x | No longer supported | Upgraded to ESLint 9 in devDependencies |
| `glob` 7.x | No longer supported | Indirect dependency |
| `q` promise library | Deprecated | Indirect dependency |

### Security Vulnerabilities

Current vulnerabilities are in development dependencies only:

1. **nth-check** (High) - RegEx complexity in SVGO
2. **postcss** (Moderate) - Parsing error 
3. **webpack-dev-server** (Moderate) - Source code exposure

**Mitigation**: These affect development tools only, not production builds.

### Production Safety

- All direct dependencies are up-to-date and secure
- Vulnerabilities are in development tools, not runtime code
- Production builds do not include vulnerable development dependencies

## Upgrade Path

To fully resolve warnings, consider:

1. **Migrate to Vite** - Modern build tool with updated dependencies
2. **Upgrade to React Scripts 6+** - When available with updated deps
3. **Custom Webpack config** - Eject and manually update dependencies

## Reporting Security Issues

Please report security vulnerabilities privately to the maintainers.