# Skipped Items

Items that were moved from TODO.md because they are:
- Too specific to individual environments
- Not suitable for a general-purpose CLI tool
- Better handled by other tools or manual configuration

## IDE Configuration

### JetBrains PyCharm Filetypes Fix
- **Original task**: `~/.config/JetBrains/PyCharm2024.3/settingsSync/options/filetypes.xml`
- **Description**: Add `<mapping pattern="Dockerfile.*" type="Dockerfile" />` to filetypes.xml
- **Reason for skipping**: This is a very specific fix for a particular PyCharm installation issue. It's too environment-specific and version-specific to be included as a general-purpose CLI command.
- **Alternative solution**: Users should manually add this mapping to their PyCharm filetypes configuration or use PyCharm's built-in settings sync features.