# gz pm doctor

Run diagnostic checks for package manager configuration and conflicts.

## Usage
```bash
gz pm doctor [flags]
```

## Flags
- `--managers <csv>`: managers to check (default: asdf)
- `--compat <auto|strict|off>`: compatibility mode
- `--check-conflicts`: check for known conflicts (default: true)
- `--fix`: attempt to fix detected issues when possible
- `--output <text|json>`: output format

## Examples
```bash
# Check conflicts across asdf plugins
gz pm doctor --check-conflicts --managers asdf

# Strict mode: fail if conflicts are detected
gz pm doctor --check-conflicts --compat strict --output json

# Attempt fixes (e.g., corepack enable)
gz pm doctor --check-conflicts --fix --managers asdf
```

## Output
- Text: sections per manager, plugin-wise warnings/conflicts/suggestions
- JSON: structured report with managers/plugins, conflicts count, warnings, suggestions, errors
