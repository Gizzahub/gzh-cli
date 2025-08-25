<!-- 🚫 AI_MODIFY_PROHIBITED -->

<!-- This file should not be modified by AI agents -->

# Manual Page Generation Specification

## Overview

The `man` command provides manual page generation capabilities for the gzh-cli CLI tool. It automatically generates Unix-style manual pages in roff format from Cobra command definitions, enabling seamless integration with system documentation standards and man page infrastructure.

## Commands

### Core Manual Generation Command

- `gz man` - Generate command line manual pages

### Manual Page Generation (`gz man`)

**Purpose**: Generate Unix-style manual pages from command definitions

**Features**:

- Automatic roff format generation
- Cobra command integration
- Hierarchical command documentation
- Standard Unix manual page structure
- Section 1 manual pages (user commands)
- Hidden command (not shown in help)

**Usage**:

```bash
gz man > gzh-manager.1                     # Generate manual page
gz man | man -l -                          # View generated manual immediately
gz man > /usr/local/share/man/man1/gz.1    # Install system-wide manual
```

**Output**: Standard roff format suitable for Unix man page systems

## Manual Page Structure

### Generated Content

The manual pages include the following standard sections:

#### NAME

- Command name and brief description
- Standard format: "gz - GZH Manager command line tool"

#### SYNOPSIS

- Command syntax and usage patterns
- All available command forms and options
- Hierarchical subcommand structure

#### DESCRIPTION

- Detailed command description
- Feature overview
- Use case explanations

#### OPTIONS

- Complete flag and option documentation
- Default values and constraints
- Required vs optional parameters

#### COMMANDS

- Subcommand documentation
- Command hierarchy representation
- Cross-references between related commands

#### EXAMPLES

- Practical usage examples
- Common workflow demonstrations
- Real-world scenarios

#### SEE ALSO

- Related commands and utilities
- External documentation references
- Additional resources

### Roff Format

The generated manual pages use standard roff markup:

- `.TH` - Manual page header
- `.SH` - Section headers
- `.TP` - Paragraph with hanging tag
- `.B` - Bold text
- `.I` - Italic text
- `.BR` - Bold-roman alternating text

## Integration

### Cobra Command Integration

The man command uses the `muesli/mango-cobra` library to:

- Extract command documentation from Cobra structures
- Generate comprehensive command descriptions
- Include all flags and subcommands
- Maintain command hierarchy structure

### System Integration

Generated manual pages integrate with:

- Unix man page system
- System manual page directories
- Man page indexing (makewhatis, mandb)
- Documentation distribution systems

## Installation and Distribution

### System Installation

```bash
# Generate and install manual page
gz man > /tmp/gz.1
sudo cp /tmp/gz.1 /usr/local/share/man/man1/
sudo mandb  # Update man page database

# Verify installation
man gz
```

### Package Distribution

Manual pages can be included in:

- Debian packages (debian/manpages)
- RPM packages (SOURCES)
- Homebrew formulas
- Distribution-specific packages

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Generate manual pages
  run: |
    ./gz man > docs/gz.1

- name: Validate manual page
  run: |
    man -l docs/gz.1 > /dev/null
```

## Output Format

### Standard Roff Structure

```roff
.TH GZ 1 "2025-01-14" "gzh-cli" "User Commands"
.SH NAME
gz \- GZH Manager command line tool
.SH SYNOPSIS
.B gz
[\fIOPTIONS\fR] \fICOMMAND\fR [\fIARGS\fR]
.SH DESCRIPTION
The GZH Manager provides comprehensive development environment management...
.SH OPTIONS
.TP
\fB\-h, \-\-help\fR
Show help for command
.TP
\fB\-\-verbose\fR
Enable verbose output
.SH COMMANDS
.TP
\fBdev-env\fR
Development environment management
.TP
\fBgit\fR
Git unified command interface
...
```

### Section Organization

Manual pages are organized into logical sections:

1. **Header information** - Command name, section, date
1. **Name and synopsis** - Basic command identification
1. **Description** - Comprehensive command overview
1. **Global options** - Common flags and parameters
1. **Subcommands** - Detailed subcommand documentation
1. **Examples** - Practical usage demonstrations
1. **Files** - Configuration files and locations
1. **Environment** - Environment variables
1. **See also** - Related commands and resources

## Command Documentation

### Automatic Documentation

The man command automatically extracts:

- Command descriptions from Cobra `Short` and `Long` fields
- Flag documentation from flag definitions
- Subcommand hierarchy and descriptions
- Usage examples from command annotations

### Documentation Quality

Generated documentation includes:

- Consistent formatting and structure
- Complete parameter documentation
- Hierarchical command organization
- Cross-references between related commands

## Examples

### Basic Manual Generation

```bash
# Generate manual page
gz man

# View generated content
gz man | less

# Generate and save to file
gz man > gz.1
```

### Installation Examples

```bash
# Install for current user
mkdir -p ~/.local/share/man/man1
gz man > ~/.local/share/man/man1/gz.1
export MANPATH="$HOME/.local/share/man:$MANPATH"

# Install system-wide (requires root)
sudo gz man > /usr/local/share/man/man1/gz.1
sudo mandb

# Verify installation
which gz
man gz
```

### Development Workflow

```bash
# Generate during development
make man-pages  # Custom build target

# Test manual page
gz man | man -l -

# Validate format
gz man | groff -man -Tascii | head -20

# Package for distribution
gz man > packaging/gz.1
```

### Documentation Updates

```bash
# After command changes
gz man > docs/gz.1
git add docs/gz.1
git commit -m "docs: update manual page"

# Verify documentation accuracy
man -l docs/gz.1
```

## Quality Assurance

### Validation

Manual pages can be validated using:

- `groff` for syntax checking
- `man -l` for rendering verification
- `mandoc -T lint` for standards compliance
- Automated testing in CI/CD

### Standards Compliance

Generated manual pages comply with:

- POSIX manual page standards
- GNU manual page conventions
- Linux manual page guidelines
- BSD manual page formats

### Testing

```bash
# Syntax validation
gz man | groff -man -Tascii > /dev/null

# Rendering test
gz man | man -l - > /dev/null

# Format verification
gz man | mandoc -T lint
```

## Integration with Build Systems

### Makefile Integration

```makefile
.PHONY: man-pages
man-pages:
	./gz man > docs/gz.1

install-man: man-pages
	install -D docs/gz.1 $(DESTDIR)/usr/share/man/man1/gz.1

clean-man:
	rm -f docs/gz.1
```

### Go Module Integration

```go
//go:generate go run main.go man > docs/gz.1
```

### Package Management

#### Debian Packaging

```bash
# debian/rules
override_dh_auto_install:
	dh_auto_install
	./gz man > debian/gz/usr/share/man/man1/gz.1
	gzip -9 debian/gz/usr/share/man/man1/gz.1
```

#### RPM Packaging

```spec
%install
./gz man > %{buildroot}%{_mandir}/man1/gz.1

%files
%{_mandir}/man1/gz.1*
```

## Error Handling

### Common Issues

- **Command structure changes**: Manual regeneration required
- **Formatting errors**: Roff syntax validation needed
- **Installation permissions**: Proper file system access required
- **Path conflicts**: Manual page location conflicts

### Error Recovery

- **Documentation drift**: Regular regeneration and validation
- **Format issues**: Automated syntax checking
- **Installation problems**: Permission and path verification
- **Compatibility issues**: Multi-platform testing

## Best Practices

### Documentation Maintenance

- Regenerate manual pages after command changes
- Validate generated content before distribution
- Include manual page generation in CI/CD pipelines
- Test manual page installation procedures

### Distribution

- Include manual pages in release packages
- Provide installation instructions for users
- Support multiple installation methods
- Maintain compatibility across Unix systems

### Development

- Keep command documentation up to date
- Use descriptive command and flag descriptions
- Include practical examples in command definitions
- Test manual page generation regularly

## Future Enhancements

### Planned Features

- **Multi-language support**: Internationalized manual pages
- **Custom templates**: Customizable manual page templates
- **Enhanced formatting**: Rich text and cross-references
- **Online documentation**: Web-based manual page viewing

### Integration Improvements

- **Help system integration**: Seamless help and manual coordination
- **Documentation testing**: Automated documentation validation
- **Content management**: Centralized documentation management
- **Format extensions**: Additional output formats (HTML, PDF)

## Security Considerations

### File System Access

Manual page generation and installation require:

- Read access to command definitions
- Write access to manual page directories
- Appropriate file system permissions

### Distribution Security

- Manual pages should be signed for package distribution
- Installation procedures should validate file integrity
- System-wide installation requires administrator privileges

## Performance Considerations

### Generation Speed

- Manual page generation is fast and lightweight
- Suitable for inclusion in build processes
- Minimal resource requirements

### System Impact

- Generated manual pages are small and efficient
- Standard roff format ensures compatibility
- Minimal system resource usage for viewing
