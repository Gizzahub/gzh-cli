# gosec suppression inventory

## 목적과 기준

이 문서는 standalone gosec suppression의 현재 상태를 기록한다. standalone
scanner는 `.make/tools.mk`에 고정된 gosec `v2.28.0`이다. TASK-116 Stage A0에서는
관측 전용 inventory였으나, Batch A에서 accepted-risk registry가 도입되면서
active directive 목록은 `security/accepted-risks.yaml`과 1:1로 연결된다.

`.gosec.json`의 `global.nosec: false` 때문에 표준 `#nosec` comment는 standalone
scan에 적용되지 않는다. 반면 `//gosec:disable`은 standalone scanner가 해석하는
active directive다. `//nolint:gosec`와 `.golangci.yml`의 gosec exclusion은
golangci-lint 표면이며 standalone 결과를 억제하지 않는다.

## Trusted base

| File | 역할 |
| ---- | ---- |
| `security/policy.yaml` | 승인 권한(GitHub numeric user id), 허용 evidence 형식, review cadence |
| `security/accepted-risks.yaml` | site 하나당 immutable `AR-YYYY-NNN` record 하나 |
| `security/internal/acceptedrisk` | 두 파일과 source directive를 fail-closed로 검증 |

승인자는 renameable한 login이 아니라 **immutable numeric user id**로 대조한다.
`type: Bot`과 모든 agent identity는 승인할 수 없다. 허용되는 evidence는
`signed-commit` 하나뿐이며, 서명이 검증되는 40자리 소문자 hex commit id여야
한다. GitHub Issue/PR URL은 사후 편집이 가능하므로 의도적으로 거부한다.

cadence: `review_by = last_reviewed_at + 90일`(저장하지 않고 파생),
hard sunset = `created_at + 180일`. hard sunset 이후에는 해당 AR ID를 갱신할 수
없으며, 새 risk 분석·새 승인·새 AR ID가 필요하다.

## Standalone active directives: 6

각 directive는 `//gosec:disable Gxxx -- AR-YYYY-NNN <reason>` 형식이며, registry
record 하나만 참조한다. 이 표는 승인 목록이 아니라 연결 상태 목록이다.

| Path                                | Line | Rule | AR ID       | Approval state         |
| ----------------------------------- | ---: | ---- | ----------- | ---------------------- |
| `cmd/selfupdate/selfupdate.go`      |  130 | G304 | AR-2026-001 | pending-owner-approval |
| `cmd/selfupdate/selfupdate.go`      |  183 | G302 | AR-2026-002 | pending-owner-approval |
| `cmd/selfupdate/selfupdate_test.go` |   33 | G304 | AR-2026-003 | pending-owner-approval |
| `cmd/dev-env/ssh_enhanced.go`       |  239 | G117 | AR-2026-004 | pending-owner-approval |
| `cmd/dev-env/ssh_enhanced.go`       | 1048 | G117 | AR-2026-005 | pending-owner-approval |
| `internal/workerpool/pool.go`       |   72 | G118 | AR-2026-006 | pending-owner-approval |

## Legacy `#nosec` comments: 0

Stage A0에서 관측된 9개의 inactive `#nosec` comment는 모두 제거되었다.
`global.nosec: false`에서 이들은 아무것도 억제하지 않으면서 "처리된 finding"이라는
잘못된 인상을 주었기 때문이다. 어느 것도 일괄 활성화하지 않았고, AR ID를 부여하지
않았다. site별 처리는 다음과 같다.

| Path                                             | Line | Rule | 처리                                                              |
| ------------------------------------------------ | ---: | ---- | ----------------------------------------------------------------- |
| `internal/pm/bootstrap/version_managers.go`      |   58 | G204 | marker 삭제. env 유래 경로를 `bash -c`에 보간하므로 finding 유지   |
| `internal/pm/bootstrap/version_managers.go`      |   95 | G204 | marker 삭제. 위와 동일                                             |
| `internal/pm/bootstrap/version_managers.go`      |  336 | G204 | marker 삭제. 위와 동일                                             |
| `internal/pm/bootstrap/version_managers.go`      |  377 | G204 | marker 삭제. 위와 동일                                             |
| `internal/pm/sync/nvm_npm.go`                    |  191 | G204 | marker 삭제. `bash -c` 보간이므로 finding 유지                     |
| `internal/pm/sync/nvm_npm.go`                    |  236 | G204 | marker 삭제 후 argument vector 사용을 설명하는 주석으로 대체       |
| `internal/pm/sync/pyenv_pip.go`                  |  193 | G204 | marker 삭제. `bash -c` 보간이므로 finding 유지                     |
| `internal/synclone/discovery/repo_discoverer.go` |  367 | G204 | marker 삭제 후 remote name 검증과 argument vector 설명 주석으로 대체 |
| `pkg/gzhclient/client.go`                        |  589 | G704 | marker 삭제 후 origin 제한을 설명하는 주석으로 대체                |

`#nosec`을 새로 추가하지 않는다. 억제가 필요하면 accepted-risk 절차를 따른다.
절차는 [security scanning 문서](../70-deployment/73-security-scanning.md#accepted-risks-the-replacement-for-nosec)에 있다.

## golangci-lint advisory surfaces

### Inline `//nolint:gosec`: 39 occurrences in 19 files

These directives are advisory only for golangci-lint. They are not a
standalone gosec suppression mechanism.

| Path                                              | Count |
| ------------------------------------------------- | ----: |
| `cmd/dev-env/aws_credentials.go`                  |     5 |
| `cmd/dev-env/azure_subscription.go`               |     2 |
| `cmd/repo-config/cmd_risk.go`                     |     1 |
| `cmd/repo-config/risk/risk.go`                    |     1 |
| `internal/auth/validator.go`                      |     1 |
| `internal/env/constants.go`                       |     3 |
| `internal/errors/recovery.go`                     |     2 |
| `internal/httpclient/secure_client.go`            |     1 |
| `internal/pm/sync/pyenv_pip.go`                   |     1 |
| `internal/pm/upgrade/homebrew.go`                 |     1 |
| `internal/simpleprof/profiler.go`                 |     3 |
| `pkg/cloud/config.go`                             |     1 |
| `pkg/config/migration.go`                         |     1 |
| `pkg/config/parser.go`                            |     2 |
| `pkg/config/repo_config_schema.go`                |     1 |
| `pkg/config/startup_validator.go`                 |     5 |
| `pkg/github/largescale/large_scale_operations.go` |     3 |
| `pkg/types/repoconfig/schema.go`                  |     1 |
| `test/e2e/helpers/cli.go`                         |     4 |

### `.golangci.yml` gosec configuration

The configuration contains 9 rule exclusions (`G104`, `G204`, `G304`, `G306`,
`G301`, `G101`, `G702`, `G703`, and `G118`) at lines 136–146. It also excludes
gosec for two broad paths: `_test\.go` at line 258 and `cmd/` at line 399.
These are golangci-lint-only exclusions; they do not change the standalone
gosec policy governed by `.gosec.json` and `GOSEC_SCAN_FLAGS`.

## 남은 blocker

`security/policy.yaml`이 등록되면서 Stage A0의 "policy owner 미등록" blocker는
해소되었다. 남은 blocker는 개별 record 승인이다. 여섯 record 모두 `evidence.sha`가
sentinel `pending-owner-approval`이며, validator가 이를 거부하므로 registry는
아직 통과하지 않는다. 이는 의도된 상태다. 실제 서명 승인 commit의 SHA로만
sentinel을 교체할 수 있다.

## Refresh procedure

Run these read-only commands from the repository root and update the counts and
paths together in a reviewed change:

```bash
rg -n -i -g '*.go' '//\s*gosec:disable'
rg -n -i -g '*.go' '#nosec'
rg -i -g '*.go' -c '//\s*nolint:gosec'
rg -n 'G[0-9]{3}|gosec' .golangci.yml
rg -n -B 8 -A 5 '^[[:space:]]*-[[:space:]]+gosec$' .golangci.yml
go test ./security/internal/acceptedrisk/...
```
