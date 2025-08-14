# Documentation Guide

This guide explains the documentation structure and management rules for the gzh-cli project.

## Documentation Categories

### 1. Core Project Documents (Protected)
**Location**: Project root directory
**Protection**: AI modification prohibited
**Files**:
- README.md - Project overview and quick start
- TECH_STACK.md - Technology stack and architecture
- FEATURES.md - Feature list and capabilities
- USAGE.md - Detailed usage instructions
- CHANGELOG.md - Version history and changes
- SECURITY.md - Security policy and practices
- LICENSE - Project license
- CLAUDE.md - AI agent instructions

**Management Rules**:
- These files contain `<!-- 🚫 AI_MODIFY_PROHIBITED -->` header
- AI agents should NOT modify these files
- Only human maintainers should update these documents
- Changes require careful review and approval

### 2. Auto-generated API Documentation
**Location**: `/api-docs/`
**Protection**: No manual edits allowed
**Content**: API reference documentation generated from code

**Management Rules**:
- Generated automatically by documentation tools
- Manual edits will be overwritten
- Source code comments should be updated instead
- Regenerate using appropriate build commands

### 3. Core Design Specifications (Protected)
**Location**: `/specs/`
**Protection**: AI modification prohibited
**Files**:
- common.md - Common functionality specifications
- dev-env.md - Development environment management specs
- net-env.md - Network environment management specs
- package-manager.md - Package manager integration specs
- synclone.md - Repository synchronization specs

**Management Rules**:
- These files contain `<!-- 🚫 AI_MODIFY_PROHIBITED -->` header
- Human-written design documents
- AI agents should NOT modify these files
- Changes require architecture review

### 4. General Documentation
**Location**: `/docs/`
**Protection**: AI modifications allowed
**Content**: Guides, tutorials, examples, and supplementary documentation

**Management Rules**:
- AI agents CAN modify these files
- Improvements and updates are welcome
- Should maintain consistency with core documents
- Regular review for accuracy

## File Protection Mechanisms

### 1. .claudeignore File
Lists files and directories that should not be modified by AI agents:
```
# AI 수정 금지 파일들
README.md
TECH_STACK.md
FEATURES.md
USAGE.md
CHANGELOG.md
SECURITY.md
LICENSE
CLAUDE.md
api-docs/**
specs/**
```

### 2. Protection Headers
Protected files include the following header:
```html
<!-- 🚫 AI_MODIFY_PROHIBITED -->
<!-- This file should not be modified by AI agents -->
```

### 3. Directory Structure
Clear separation between protected and editable documentation:
```
/                    # Protected core documents
├── api-docs/        # Auto-generated (no manual edits)
├── specs/           # Protected design specs
└── docs/            # General documentation (AI editable)
```

## Documentation Standards

### Markdown Formatting
- Use consistent heading levels
- Include table of contents for long documents
- Use code blocks with language identifiers
- Add examples and use cases

### File Naming
- Use lowercase with hyphens: `feature-name.md`
- Descriptive names that indicate content
- README.md for directory overviews

### Content Guidelines
- Keep documentation up-to-date with code changes
- Include practical examples
- Link to related documentation
- Use clear, concise language

### Version Control
- Commit documentation changes separately
- Use descriptive commit messages
- Review documentation changes carefully
- Maintain documentation history

## Adding New Documentation

### For Protected Categories (1 & 3)
1. Human maintainer creates the file
2. Add protection header if needed
3. Update .claudeignore if necessary
4. Review and approve changes

### For General Documentation (Category 4)
1. Create file in `/docs/` directory
2. Follow documentation standards
3. AI agents can help improve content
4. Regular review for accuracy

### For API Documentation (Category 2)
1. Update source code comments
2. Run documentation generation tools
3. Commit generated files
4. Do not edit generated files directly

## Documentation Review Process

1. **Regular Reviews**: Schedule periodic documentation reviews
2. **Code-Documentation Sync**: Ensure docs match implementation
3. **User Feedback**: Incorporate user suggestions
4. **Consistency Checks**: Maintain consistent style and structure

## Best Practices

1. **Write for Your Audience**: Consider who will read the documentation
2. **Keep It Current**: Update docs with code changes
3. **Be Comprehensive**: Cover all features and use cases
4. **Stay Organized**: Use clear structure and navigation
5. **Include Examples**: Show, don't just tell
6. **Test Documentation**: Verify examples and instructions work

## Common Documentation Tasks

### Updating General Documentation
```bash
# AI agents can help with these files
docs/getting-started.md
docs/troubleshooting.md
docs/examples/*.md
```

### Reviewing Protected Documentation
```bash
# Only human maintainers should modify
README.md
TECH_STACK.md
specs/*.md
```

### Generating API Documentation
```bash
# Use appropriate tools
make generate-docs  # Example command
```

## Documentation Maintenance

1. **Broken Links**: Check and fix regularly
2. **Outdated Information**: Update with each release
3. **Missing Documentation**: Identify and fill gaps
4. **Consistency**: Maintain style and format
5. **Accessibility**: Ensure documentation is easy to find and read

This guide ensures consistent, high-quality documentation across the project while maintaining clear boundaries for AI-assisted contributions.
