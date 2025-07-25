# GZH Manager v2.0 Rollback Plan

## Overview
This document outlines procedures for rolling back the command consolidation changes if critical issues arise.

## Rollback Triggers

### Severity Levels

#### Critical (Immediate Rollback)
- Data loss or corruption
- Security vulnerabilities introduced
- Complete feature breakage affecting >30% users
- Performance degradation >50%

#### High (Rollback within 24 hours)
- Major features non-functional
- Migration script failures affecting >20% users
- CI/CD breakage for enterprise customers

#### Medium (Evaluate rollback)
- Multiple minor bugs reported
- Performance degradation 20-50%
- Documentation significantly incorrect

#### Low (Fix forward)
- Minor bugs
- UI/UX complaints
- Documentation typos

## Rollback Strategies

### Strategy 1: Feature Flag Rollback (Preferred)
**Time to rollback**: 5 minutes

```go
// internal/features/flags.go
var CommandConsolidation = FeatureFlag{
    Name: "command_consolidation",
    Default: true,
}

// To rollback:
// 1. Set flag to false in config service
// 2. Users automatically get old commands
```

**Implementation**:
```bash
# Emergency rollback
curl -X POST https://api.gz.dev/flags/command_consolidation \
  -d '{"enabled": false}' \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Gradual rollback
curl -X POST https://api.gz.dev/flags/command_consolidation \
  -d '{"enabled": false, "percentage": 50}'
```

### Strategy 2: Version Rollback
**Time to rollback**: 30 minutes

```bash
# For users who installed via package manager
brew install gzh-manager@1.9.0

# For users who built from source
git checkout v1.9.0
make install

# For Docker users
docker pull gizzahub/gz:1.9.0
```

### Strategy 3: Alias-Based Compatibility
**Time to rollback**: Immediate

```bash
# Enable full compatibility mode
export GZH_COMPATIBILITY_MODE=v1

# This makes new commands behave exactly like old ones
gz gen-config  # Works without warnings
```

## Rollback Procedures

### Phase 1: Detection & Decision (0-2 hours)
1. **Monitor alerts**:
   - Error rate spike detection
   - User complaint threshold
   - Performance degradation alerts

2. **Assess impact**:
   - Number of affected users
   - Severity of issues
   - Business impact

3. **Make decision**:
   - Rollback decision tree
   - Stakeholder approval if needed

### Phase 2: Execute Rollback (2-4 hours)

#### For Feature Flag Rollback:
```bash
# 1. Disable feature flag
./scripts/rollback.sh --feature-flag

# 2. Verify rollback
./scripts/verify-rollback.sh

# 3. Notify users
./scripts/notify-users.sh --rollback
```

#### For Version Rollback:
```bash
# 1. Tag rollback point
git tag rollback-v2.0.0

# 2. Revert to previous version
git checkout v1.9.0
git cherry-pick <critical-fixes>

# 3. Emergency release
make release VERSION=v1.9.1-hotfix

# 4. Update package managers
./scripts/update-packages.sh --emergency
```

### Phase 3: Communication (4-6 hours)
1. **Internal communication**:
   - Engineering team notification
   - Support team briefing
   - Executive summary

2. **External communication**:
   - Status page update
   - User email notification
   - Social media announcement
   - GitHub issue creation

3. **Documentation update**:
   - Rollback notice in README
   - Known issues documentation
   - Temporary workarounds

## Rollback Testing

### Pre-Release Testing
```bash
# Test rollback procedures before release
./tests/rollback/test-feature-flag.sh
./tests/rollback/test-version-rollback.sh
./tests/rollback/test-data-integrity.sh
```

### Rollback Scenarios
1. **Scenario A**: Feature flag rollback
   - Expected time: 5 minutes
   - Data loss: None
   - User impact: Minimal

2. **Scenario B**: Version rollback
   - Expected time: 30 minutes
   - Data loss: None (config backup)
   - User impact: Need to reinstall

3. **Scenario C**: Partial rollback
   - Expected time: 1 hour
   - Data loss: None
   - User impact: Some features unavailable

## Data Protection

### Configuration Backup
```bash
# Automatic backup before migration
~/.config/gzh-manager/backups/
├── pre-v2.0.0/
│   ├── config.yaml
│   ├── metadata.json
│   └── timestamp
```

### Restore Procedures
```bash
# Restore configuration
gz-migrate --rollback

# Manual restore
cp ~/.config/gzh-manager/backups/pre-v2.0.0/* ~/.config/gzh-manager/
```

## Post-Rollback Actions

### Immediate (0-24 hours)
1. Root cause analysis
2. Fix critical issues
3. Update test coverage
4. Prepare patch release

### Short-term (1-7 days)
1. Comprehensive fix development
2. Extended beta testing
3. Updated migration plan
4. Enhanced monitoring

### Long-term (1-4 weeks)
1. Re-release planning
2. Additional safeguards
3. Process improvements
4. Documentation updates

## Rollback Communication Templates

### User Notification
```
Subject: GZH Manager v2.0 Temporary Rollback

Dear GZH Manager Users,

We've identified an issue with v2.0 and have temporarily rolled back to ensure stability. 

What this means:
- Old commands are available again
- Your workflows continue to work
- No data has been lost

Next steps:
- Continue using GZH Manager as before
- Watch for updates on the fix
- v2.0.1 will be released soon

We apologize for any inconvenience.
```

### Status Page Update
```markdown
## GZH Manager v2.0 Rollback
**Status**: Rolled Back
**Impact**: Command structure reverted to v1.x

We've rolled back v2.0 due to [issue description]. 
Old commands are working normally. 
ETA for fix: [timeframe]
```

## Success Criteria

### Successful Rollback Indicators
- Error rates return to baseline
- User complaints cease
- Performance metrics normal
- No data loss reported

### Failed Rollback Indicators
- Continued errors
- Rollback introduces new issues
- Data inconsistencies
- User confusion

## Lessons Learned Process

After any rollback:
1. Conduct blameless postmortem
2. Document root causes
3. Update rollback procedures
4. Improve testing coverage
5. Share learnings with team

## Emergency Contacts

- **Technical Lead**: [Contact]
- **Product Manager**: [Contact]
- **DevOps On-Call**: [Contact]
- **Support Lead**: [Contact]
- **Communications**: [Contact]

## Appendix: Rollback Scripts

### rollback.sh
```bash
#!/bin/bash
# Emergency rollback script

set -e

echo "Starting GZH Manager rollback..."

# Check rollback type
if [[ "$1" == "--feature-flag" ]]; then
    echo "Executing feature flag rollback..."
    # Feature flag rollback logic
elif [[ "$1" == "--version" ]]; then
    echo "Executing version rollback..."
    # Version rollback logic
fi

echo "Rollback complete"
```

This rollback plan ensures we can quickly recover from any issues while minimizing user impact.