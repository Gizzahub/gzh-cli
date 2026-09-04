# gosec suppression inventory

## 목적과 기준

이 문서는 standalone gosec suppression의 현재 상태를 기록한다. standalone
scanner는 `.make/tools.mk`에 고정된 gosec `v2.28.0`이다. TASK-116 Stage A0에서는
관측 전용 inventory였으나, Batch A에서 accepted-risk registry가 도입되면서
active directive 목록은 `security/accepted-risks.yaml`과 1:1로 연결된다.

`//gosec:disable`은 standalone scanner가 해석하는 active directive다.
`//nolint:gosec`와 `.golangci.yml`의 gosec exclusion은 golangci-lint 표면이며
standalone 결과를 억제하지 않는다.

### blanket tag는 고정값이 아니라 설정값이다

gosec은 `#nosec`을 하드코딩하지 않는다. 실행 시점에 `.gosec.json`에서 tag를
만들어낸다. pinned binary로 실측한 결과는 다음과 같다.

| `global` 설정                        | gosec이 존중하는 tag                |
| ------------------------------------ | ----------------------------------- |
| `nosec` key 없음                     | `#nosec`                            |
| `"nosec": false` (**이 저장소**)     | `#false`                            |
| `"nosec": "skipme"`                  | `#skipme`                           |
| `"nosec": false, "#nosec": "skipme"` | `#false`와 `#skipme` 둘 다          |
| `"nosec": true`                      | 없음 — `//gosec:disable`까지 무력화 |

live tag는 `"#"` + 설정값이다. 따라서 이 저장소의 `false` 설정에서 살아 있는
blanket 형식은 `#false`이고 `#nosec` 자체는 무력하다. `#nosec` key는 기본 tag를
대체하는 것이 아니라 두 번째 tag를 추가하는 별개 설정이다.

`true`는 더 강한 설정이 아니다. gosec의 `ignoreNosec`을 켜서 `//gosec:disable`을
포함한 모든 억제 문법을 끄므로, 이 registry의 directive가 억제를 멈추고 이미 승인된
risk에서 gate가 실패한다. 이 비자명한 이유로 `false`만이 유효한 값이다.

### 강제 방식

- validator는 live blanket tag를 그 자체로 위반(`suppression-blanket-form`)으로
  본다. AR record를 지목하지 않고 지목할 수도 없으므로, 파싱에 실패한 directive가
  아니라 등록 불가능한 억제다. scanner는 gosec과 동일하게 `.gosec.json`에서 tag를
  유도하며, 설정이 사라져도 침묵하지 않도록 `#nosec`을 항상 포함한다. scan 대상
  파일의 comment에서 line comment와 `/* */` block comment를 모두 읽는다. string
  literal 안의 tag는 comment가 아니므로 보고하지 않는다.
- scanner는 gosec이 매칭하는 위치에서만 매칭한다. comment 줄 시작(들여쓰기 허용),
  대소문자 구분, tag 뒤 구분자 불필요. 대문자 표기와 문장 중간 언급은 gosec이
  무시하므로 scanner도 무시한다. 그래서 이 저장소의 산문은 blanket 형식을 자유롭게
  언급할 수 있다 — live tag로 *시작하는* comment 줄만 억제다.
- `TestRepositoryGosecConfigDisablesBlanketSuppression`이 `.gosec.json`을 읽어
  `global.nosec`이 `false`가 아니게 되거나, `#nosec` alternative key가 추가되거나,
  유도된 tag 집합이 `#false` + `#nosec`이 아니게 되면 실패한다. 설정과 scanner의
  괴리를 막는 장치다.

## Trusted base

| File                             | 역할                                                                  |
| -------------------------------- | --------------------------------------------------------------------- |
| `security/policy.yaml`           | 승인 권한(GitHub numeric user id), 허용 evidence 형식, review cadence |
| `security/accepted-risks.yaml`   | site 하나당 immutable `AR-YYYY-NNN` record 하나                       |
| `security/internal/acceptedrisk` | 두 파일과 source directive를 fail-closed로 검증                       |

승인자는 renameable한 login이 아니라 **immutable numeric user id**로 대조한다.
`type: Bot`과 모든 agent identity는 승인할 수 없다. 허용되는 evidence는
`signed-commit` 하나뿐이다. GitHub Issue/PR URL은 사후 편집이 가능하므로
의도적으로 거부한다.

approver login은 GitHub이 실제로 발급할 수 있는 형태여야 한다. ASCII 영숫자와
내부 단일 hyphen만 허용하고, 39자 이하이며, hyphen으로 시작하거나 끝날 수 없다.
automation marker 검사는 ASCII substring 비교라서 Unicode homoglyph로 쓴 login을
전혀 보지 못한다. 존재할 수 없는 login은 권한자일 수 없으므로, 문자 규칙이 marker
목록과 무관하게 그 경로를 닫는다. 다만 marker가 없는 automation 이름
(`gzh-release-automation`)은 두 규칙 모두 걸러내지 못한다. 계정이 사람의 것인지는
문자에 대한 판단이 아니라 소유에 대한 판단이므로, approver를 추가하는 review의
몫으로 남긴다.

`security/policy.yaml`과 `security/accepted-risks.yaml`은 각각 YAML document를
정확히 하나만 담아야 한다. 뒤따르는 내용은 형태가 아니라 존재로 거부한다. scalar,
sequence, 타입이 다른 mapping, unknown field를 가진 mapping은 모두 append된
document이며 어느 것도 기대 타입으로 decode되지 않으므로, decode 실패를 "뒤에
아무것도 없음"으로 읽는 guard는 이들을 전부 통과시킨다.

서명이 검증되기만 해서는 승인이 아니다. trusted base에 push할 수 있는 사람은
누구나 자기 키로 commit에 서명할 수 있기 때문이다. approval commit은 다음 세
가지를 모두 만족해야 한다.

1. 기록된 40자리 소문자 hex SHA에 대해 서명이 검증된다.
1. verifier가 서명에서 확인한 key fingerprint가 **해당 record의 approver**의
   `signing_keys`에 등록되어 있고, verifier가 그 키를 forge account로
   해석했다면 그 account id가 approver id와 같다. commit의 author/committer
   header는 인증되지 않은 문자열이므로 identity로 취급하지 않는다.
1. commit message가 승인 대상 `AR-YYYY-NNN`을 명시한다. 서명이 message를
   포함하므로, 서명 하나가 registry 전체를 조용히 승인하는 일을 막는다.

`signing_keys`는 현재 빈 목록이며, 실제 fingerprint를 임의로 만들어 넣지 않았다.
빈 목록은 "이 approver를 만족시킬 수 있는 서명이 없다"는 뜻이고, owner가 실제로
서명에 쓰는 키의 fingerprint를 등록할 때까지 유지되는 fail-closed 상태다.

scanner가 건너뛰는 directory 집합은 `.make/tools.mk`의 `GOSEC_SCAN_FLAGS`에서
파생한다. gosec이 실제로 읽는 directory보다 좁아질 수 없어야 하기 때문이다.
`build/`, `bin/`, `dist/`는 gosec이 scan하고 directive를 존중하므로 registry도
동일하게 본다.

cadence: `review_by = last_reviewed_at + 90일`(저장하지 않고 파생),
hard sunset = `created_at + 180일`. hard sunset 이후에는 해당 AR ID를 갱신할 수
없으며, 새 risk 분석·새 승인·새 AR ID가 필요하다.

`created_at`과 `last_reviewed_at`은 cadence 검사와 동일한 평가 시점을 기준으로
미래 날짜인지 확인한다(`record-date-in-future`). 미래로 적은 `last_reviewed_at`은
review 주기 한 번을 통째로 벌고, 미래로 적은 `created_at`은 hard sunset을 같은
만큼 밀어내기 때문이다.

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
`global.nosec: false`에서 live tag는 `#false`이므로 이들은 아무것도 억제하지
않으면서 "처리된 finding"이라는 잘못된 인상을 주었기 때문이다. 어느 것도 일괄 활성화하지 않았고, AR ID를 부여하지
않았다. site별 처리는 다음과 같다.

| Path                                             | Line | Rule | 처리                                                                 |
| ------------------------------------------------ | ---: | ---- | -------------------------------------------------------------------- |
| `internal/pm/bootstrap/version_managers.go`      |   58 | G204 | marker 삭제. env 유래 경로를 `bash -c`에 보간하므로 finding 유지     |
| `internal/pm/bootstrap/version_managers.go`      |   95 | G204 | marker 삭제. 위와 동일                                               |
| `internal/pm/bootstrap/version_managers.go`      |  336 | G204 | marker 삭제. 위와 동일                                               |
| `internal/pm/bootstrap/version_managers.go`      |  377 | G204 | marker 삭제. 위와 동일                                               |
| `internal/pm/sync/nvm_npm.go`                    |  191 | G204 | marker 삭제. `bash -c` 보간이므로 finding 유지                       |
| `internal/pm/sync/nvm_npm.go`                    |  236 | G204 | marker 삭제 후 argument vector 사용을 설명하는 주석으로 대체         |
| `internal/pm/sync/pyenv_pip.go`                  |  193 | G204 | marker 삭제. `bash -c` 보간이므로 finding 유지                       |
| `internal/synclone/discovery/repo_discoverer.go` |  367 | G204 | marker 삭제 후 remote name 검증과 argument vector 설명 주석으로 대체 |
| `pkg/gzhclient/client.go`                        |  589 | G704 | marker 삭제 후 origin 제한을 설명하는 주석으로 대체                  |

blanket tag를 새로 추가하지 않는다. 새로 추가하면 validator가
`suppression-blanket-form`으로 거부한다. 억제가 필요하면 accepted-risk 절차를 따른다.
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
해소되었다. 남은 blocker는 두 가지이며 모두 owner만 해소할 수 있다.

1. `approvers[].signing_keys`가 비어 있다. owner가 실제 서명 키의 fingerprint를
   등록하기 전에는 어떤 서명도 승인으로 인정되지 않는다.
1. 여섯 record 모두 `evidence.sha`가 sentinel `pending-owner-approval`이다.
   실제 서명 승인 commit의 SHA로만 교체할 수 있다.

validator가 두 조건을 모두 거부하므로 registry는 아직 통과하지 않는다. 이는
의도된 상태다.

## Refresh procedure

Run these read-only commands from the repository root and update the counts and
paths together in a reviewed change:

```bash
rg -n -i -g '*.go' '//\s*gosec:disable'
rg -n -g '*.go' '^\s*(//|/\*)?\s*#(nosec|false)'   # live blanket tags; see .gosec.json
rg -i -g '*.go' -c '//\s*nolint:gosec'
rg -n 'G[0-9]{3}|gosec' .golangci.yml
rg -n -B 8 -A 5 '^[[:space:]]*-[[:space:]]+gosec$' .golangci.yml
go test ./security/internal/acceptedrisk/...
```
