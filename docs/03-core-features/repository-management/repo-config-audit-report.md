# Repository Compliance Audit Report Guide

## Overview

The `gz repo-config audit` command generates comprehensive compliance audit reports for your GitHub organization's repositories. With the enhanced HTML report template, you can now generate beautiful, interactive reports with visualizations and filtering capabilities.

## Features

### Enhanced HTML Reports

The new HTML report template provides:

1. **Interactive Dashboard**
   - Real-time filtering by repository name, status, and policies
   - Searchable repository list
   - Print-friendly layout

2. **Visual Compliance Score**
   - Circular progress indicator showing overall compliance percentage
   - Color-coded based on compliance level:
     - Green (≥80%): Excellent compliance
     - Yellow (60-79%): Good compliance
     - Red (<60%): Needs improvement

3. **Key Metrics Cards**
   - Total repositories
   - Compliant repositories
   - Non-compliant repositories
   - Total violations

4. **Policy Overview**
   - List of active policies with enforcement levels
   - Violation counts per policy
   - Visual badges for policy types

5. **Detailed Repository Table**
   - Status indicators
   - Violation details with policy and rule information
   - Applied policies per repository
   - Last checked timestamp

6. **Compliance Trend Chart**
   - 30-day trend visualization
   - Track compliance improvements over time
   - Interactive chart with Chart.js

## Usage

### Generate HTML Report

```bash
# Basic HTML report
gz repo-config audit --org myorg --format html

# Save to file
gz repo-config audit --org myorg --format html --output compliance-report.html

# With detailed violations
gz repo-config audit --org myorg --format html --detailed --output report.html
```

### Other Formats

```bash
# Table format (console output)
gz repo-config audit --org myorg

# JSON format
gz repo-config audit --org myorg --format json

# CSV format
gz repo-config audit --org myorg --format csv --output report.csv
```

## HTML Report Structure

### 1. Header Section
- Organization name
- Report generation timestamp
- Print and export buttons

### 2. Metrics Dashboard
- Four key metric cards showing compliance statistics
- Visual compliance score with circular progress

### 3. Policy Overview
- Active policies with enforcement levels (required/recommended)
- Violation counts per policy

### 4. Filter Controls
- Search box for repository names
- Status filter (All/Compliant/Non-Compliant)
- Policy filter
- Reset button

### 5. Repository Details Table
- Repository information
- Compliance status
- Detailed violations
- Applied policies
- Last checked time

### 6. Compliance Trend Chart
- Line chart showing compliant vs non-compliant repositories over time
- 30-day historical view

## Customization

### Template Location

The HTML template is embedded in the binary but can be customized by:

1. Modifying `cmd/repo-config/templates/audit-report.html`
2. Rebuilding the binary

### Styling

The template uses:
- Bootstrap 5.3 for responsive layout
- Font Awesome 6.4 for icons
- Chart.js 4.3 for visualizations
- Custom CSS variables for theming

### Adding Custom Metrics

To add custom metrics to the report:

1. Update the `HTMLTemplateData` struct
2. Modify the template to display new metrics
3. Update the data generation logic

## Best Practices

1. **Regular Audits**
   - Schedule weekly or monthly audits
   - Track compliance trends over time
   - Set compliance targets

2. **Report Distribution**
   - Use HTML format for management presentations
   - Use JSON format for automated processing
   - Use CSV format for spreadsheet analysis

3. **Compliance Tracking**
   - Set up automated alerts for compliance drops
   - Review violation details regularly
   - Update policies based on audit findings

## Integration Examples

### CI/CD Pipeline

```yaml
# GitHub Actions example
- name: Generate Compliance Report
  run: |
    gz repo-config audit --org ${{ github.repository_owner }} \
      --format html --output audit-report.html
    
- name: Upload Report
  uses: actions/upload-artifact@v3
  with:
    name: compliance-report
    path: audit-report.html
```

### Automated Reporting

```bash
#!/bin/bash
# Generate weekly compliance report

DATE=$(date +%Y%m%d)
ORG="myorg"

# Generate report
gz repo-config audit --org $ORG --format html \
  --output "compliance-${ORG}-${DATE}.html"

# Email report (example)
mail -s "Weekly Compliance Report" team@company.com \
  < "compliance-${ORG}-${DATE}.html"
```

## Troubleshooting

### Large Organizations

For organizations with many repositories:

```bash
# Use JSON format for better performance
gz repo-config audit --org large-org --format json > audit.json

# Process in batches
gz repo-config audit --org large-org --filter "^api-.*"
```

### Template Issues

If the HTML template fails to render:
- The system falls back to a simple HTML format
- Check console output for template parsing errors
- Verify the template file is properly embedded

## Future Enhancements

Planned improvements include:
- PDF export functionality
- Email report scheduling
- Custom branding options
- More visualization types
- Historical comparison reports