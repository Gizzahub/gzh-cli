// Package gen_config provides intelligent configuration file generation
// and discovery for the GZH Manager system.
//
// This package implements the gen-config command that can automatically
// generate configuration files by discovering existing repositories and
// development environments, making it easy to bootstrap GZH Manager
// configurations for new projects or teams.
//
// Key Features:
//
// Configuration Discovery:
//   - Automatic repository discovery in directory trees
//   - Git platform detection (GitHub, GitLab, Gitea)
//   - Existing configuration analysis
//   - Development pattern recognition
//
// Template System:
//   - Flexible configuration templates
//   - Context-aware template selection
//   - Customizable generation rules
//   - Multi-format output support (YAML, JSON)
//
// Generation Modes:
//   - Interactive mode with user prompts
//   - Automatic mode with intelligent defaults
//   - Template-based generation
//   - Incremental configuration updates
//
// Configuration Types:
//   - Bulk clone configurations
//   - Monitoring setups
//   - Network environment profiles
//   - IDE settings templates
//   - CI/CD pipeline configurations
//
// Example usage:
//
//	gz gen-config --discover /path/to/projects
//	gz gen-config --template github-org --output bulk-clone.yaml
//	gz gen-config --interactive --type monitoring
//	gz gen-config --update-existing config.yaml
//
// The package helps teams quickly set up GZH Manager configurations
// by analyzing existing development environments and generating
// appropriate configuration files with minimal manual intervention.
package genconfig
