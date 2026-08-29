# gosec suppression inventory

## 목적과 기준

이 문서는 TASK-116 Stage A0의 관측 전용 inventory다. 기준 revision은
`3fce0878a1bd3d13cda8ebbe14d2ee2fe001ce67`이며, standalone scanner는
`.make/tools.mk`에 고정된 gosec `v2.28.0`이다. 이 문서는 suppression을
승인하거나, 활성화하거나, source/configuration을 바꾸지 않는다.

`.gosec.json`의 `global.nosec: false` 때문에 표준 `#nosec` comment는
standalone scan에 적용되지 않는다. 반면 `//gosec:disable`은 standalone
scanner가 해석하는 active directive다. `//nolint:gosec`와 `.golangci.yml`의
gosec exclusion은 golangci-lint 표면이며 standalone 결과를 억제하지 않는다.

## Standalone active directives: 6

아래 여섯 directive에는 현재 accepted-risk ID(AR ID)가 없다. 존재 사실은
기록하되, 이 목록은 승인 목록이 아니다.

| Path                                | Line | Rule |
| ----------------------------------- | ---: | ---- |
| `cmd/dev-env/ssh_enhanced.go`       |  239 | G117 |
| `cmd/dev-env/ssh_enhanced.go`       | 1048 | G117 |
| `cmd/selfupdate/selfupdate.go`      |  130 | G304 |
| `cmd/selfupdate/selfupdate.go`      |  183 | G302 |
| `cmd/selfupdate/selfupdate_test.go` |   33 | G304 |
| `internal/workerpool/pool.go`       |   72 | G118 |

향후 승인된 standalone 예외의 유일한 source 형식은 다음과 같다. `AR-ID`는
policy owner가 immutable approval evidence에 연결하여 발급한 값이어야 하며,
예시는 실제 ID가 아니다.

```go
//gosec:disable Gxxx -- AR-ID: approved, specific reason.
```

## Inactive legacy `#nosec` comments: 9

이 아홉 comment는 `nosec: false` 설정에서 standalone scan을 억제하지 않는다.
자동 변환, 활성화, AR ID 부여는 이 Stage의 범위 밖이다.

| Path                                             | Line | Rule |
| ------------------------------------------------ | ---: | ---- |
| `pkg/gzhclient/client.go`                        |  589 | G704 |
| `internal/synclone/discovery/repo_discoverer.go` |  367 | G204 |
| `internal/pm/sync/pyenv_pip.go`                  |  193 | G204 |
| `internal/pm/sync/nvm_npm.go`                    |  191 | G204 |
| `internal/pm/sync/nvm_npm.go`                    |  236 | G204 |
| `internal/pm/bootstrap/version_managers.go`      |   58 | G204 |
| `internal/pm/bootstrap/version_managers.go`      |   95 | G204 |
| `internal/pm/bootstrap/version_managers.go`      |  336 | G204 |
| `internal/pm/bootstrap/version_managers.go`      |  377 | G204 |

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

## Policy-owner blocker

No policy owner identity or immutable approval-evidence format is registered in
the repository. Consequently, no one may ratify the six active directives,
convert legacy `#nosec` comments, or issue AR IDs under TASK-116 yet. The policy
owner must first define the authoritative identity and evidence location; only
then can a later task establish the accepted-risk registry and enforcement.

## Refresh procedure

Run these read-only commands from the repository root and update the counts and
paths together in a reviewed change:

```bash
rg -n -i -g '*.go' '//\s*gosec:disable'
rg -n -i -g '*.go' '#nosec'
rg -i -g '*.go' -c '//\s*nolint:gosec'
rg -n 'G[0-9]{3}|gosec' .golangci.yml
rg -n -B 8 -A 5 '^[[:space:]]*-[[:space:]]+gosec$' .golangci.yml
```
