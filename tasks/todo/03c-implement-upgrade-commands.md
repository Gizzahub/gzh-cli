# TODO: 업그레이드 CLI 명령어 및 매니저 구현

---
status: [ ]
priority: medium
severity: medium
file_type: service_layer
estimated_effort: 30분
source: 03-implement-pm-upgrade-managers.md (분할됨)
depends_on: [03b-implement-upgrade-managers.md]
spec_reference: `cmd/pm/advanced.go:71`
---

## 📋 작업 개요

`gz pm upgrade-managers` 명령어의 실제 구현과 업그레이드 매니저 조정자를 완성합니다. 현재 "not yet implemented" 상태를 완전한 기능으로 대체합니다.

## 🎯 구현 목표

### Step 1: 업그레이드 매니저 구현
백업, 롤백, 버전 비교 등의 핵심 기능을 포함한 매니저를 구현합니다.

```go
type UpgradeManager struct {
    upgraders map[string]PackageManagerUpgrader
    logger    logger.Logger
    backupDir string
}

func (um *UpgradeManager) CheckAll(ctx context.Context) (*UpgradeReport, error)
func (um *UpgradeManager) UpgradeManagers(ctx context.Context, names []string, opts UpgradeOptions) (*UpgradeReport, error)
```

### Step 2: CLI 명령어 완성
`cmd/pm/advanced.go`의 upgrade-managers 명령어를 실제 구현으로 교체합니다.

```bash
# 지원할 명령어 형식
gz pm upgrade-managers --check
gz pm upgrade-managers --all
gz pm upgrade-managers --manager brew,nvm
gz pm upgrade-managers --all --backup
```

## 📁 파일 구조

### 생성할 파일
- `internal/pm/upgrade/manager.go` - 업그레이드 매니저 구현
- `internal/pm/upgrade/version_comparator.go` - 버전 비교 유틸리티
- `internal/pm/upgrade/backup.go` - 백업 및 롤백 로직

### 수정할 파일
- `cmd/pm/advanced.go` - upgrade-managers 명령어 실제 구현

## ✅ 완료 기준

- [ ] 백업/롤백 기능 안정성
- [ ] 에러 상황 적절한 처리

## 🚀 커밋 메시지

```
feat(claude-opus): 업그레이드 CLI 명령어 및 매니저 완성

- 업그레이드 매니저 조정자 구현
- 백업/롤백 시스템 통합
- upgrade-managers 명령어 완전 구현
- 버전 비교 및 에러 처리 로직

Closes: cmd/pm/advanced.go:71 "upgrade-managers command not yet implemented"

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```
