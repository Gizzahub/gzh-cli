# 출시 산출물·서명 전달 계약

## 적용 범위

첫 공개 후보는 사용자 승인된 `v0.2.0`이다. 소스 통합이나 snapshot 성공이 태그·서명·게시
승인을 대신하지 않는다. 공개 채널은 GitHub Release와 기존 `Gizzahub/homebrew-tap`의
`Casks/gz.rb`다. Docker/GHCR, NFPM 패키지, Scoop, AUR, Chocolatey와 알림은 이번 채널에
포함하지 않는다. 라이선스 metadata는 실제 LICENSE·README와 같은 MIT다.

GoReleaser는 archive·SBOM·cask를 만들고 게시하지 않는다. `.goreleaser.yml`의
`release.disable: true`와 cask의 `skip_upload: true`는 정상 GoReleaser 명령을 실행해도
원격 게시가 일어나지 않게 한다. 실제 게시자는 아래의 검증을 통과한 기존 bytes만 업로드한다.

## 단계와 권한

| 단계             | 권한                                | 수행 내용                                                                      |
| ---------------- | ----------------------------------- | ------------------------------------------------------------------------------ |
| 수동 branch 검증 | contents: read                      | 고정 도구 검사, offline 회귀, unsigned snapshot과 전달 구조 확인               |
| 승인된 tag push  | contents: read, id-token: write     | 동일 소스의 archive/SBOM/cask, 최종 manifest/checksum 생성과 keyless 서명·검증 |
| 승인된 로컬 게시 | infra-ops의 저장소별 단기 App token | 원격 tag·소스·서명·파일 일치를 확인하고 새 Release와 tap만 게시                |

수동 dispatch에는 OIDC나 게시 credential이 없다. tag job에도 contents/packages write나
장기 게시 secret이 없다. keyless 서명은 Fulcio/Rekor의 공개 기록을 남기므로 실제 tag 승인
후의 hosted job에서만 실행한다. 로컬 snapshot·테스트는 서명하지 않는다.

## 산출물 전달

`scripts/releaseartifact`가 GoReleaser v2.19의 `metadata.json`과 `artifacts.json`을 읽어
새 `dist/publish` 디렉터리를 만든다. 기존 stage가 있으면 덮어쓰지 않는다.

- Linux amd64·arm64·armv6·armv7, macOS amd64·arm64, Windows amd64의 archive 7개
- 각 archive에 대응하는 SBOM 7개
- `gz.rb`, `release-notes.md`, `release-manifest.json`
- 위 파일 전부를 포함하는 `checksums.txt`
- tag job이 최종 checksum에 대해 생성한 `checksums.txt.sigstore.json`

manifest는 tag·version·정확한 source SHA·snapshot 여부와 asset별 이름·크기·SHA256·
플랫폼/대응 archive를 고정한다. cask의 4개 URL/hash는 같은 archive와 일치해야 한다.
archive에는 최상위 `gz` 또는 `gz.exe`, 정확한 source의 LICENSE·README·CHANGELOG,
기존 이름의 completions 3개와 manpage가 필요하다. archive 경로 탈출·중복·link/device,
symlink stage·과도한 크기·불완전한 inventory는 거부한다.

snapshot은 metadata의 SNAPSHOT version과 별도 표시된 본문으로 구분한다. snapshot의
파일 구조 검증 성공은 정식 signature·release 성공이 아니다. verify/publish는 snapshot을
거부한다. 안정 tag의 본문은 생성된 dist CHANGELOG가 있어야 한다.

현재 validator의 SBOM 검사는 JSON 형식·archive 연결·signed hash 일치이며 전체 SPDX
schema 준수를 증명하지 않는다. binary member도 존재·크기·archive hash를 검증하며
Go buildinfo의 독립 판정은 별도 실제 산출물 검사에 속한다. 최종 출시에서는 플랫폼별
실제 payload/buildinfo, 라이선스, SBOM 내용과 설치를 추가 확인한다.

## 서명 검증

서명은 모든 파일을 stage에 확정한 뒤 최종 checksums에 한 번 수행한다. publisher는
산출물을 다시 빌드하거나 checksum·manifest를 재생성하지 않는다.

cosign 검증은 고정 v3.1.3 실행 파일과 다음 identity 전체를 요구한다.

- identity: `https://github.com/Gizzahub/gzh-cli/.github/workflows/release.yml@refs/tags/<tag>`
- issuer: `https://token.actions.githubusercontent.com`
- workflow SHA: 승인된 source SHA
- workflow repository: `Gizzahub/gzh-cli`
- workflow ref: `refs/tags/<tag>`
- workflow trigger: `push`

repository 전체에 대한 느슨한 regexp나 insecure CT/tlog 예외를 쓰지 않는다.
검증은 stage에서 확보한 정확한 checksum·bundle bytes에 결합해야 하며, 검사 도중 원본
경로를 교체·복원하는 것으로 다른 파일의 서명을 빌릴 수 없어야 한다. 게시에도 검증한
동일한 메모리 bytes를 사용한다.

근거: [cosign v3.1.3 verify-blob 계약](https://github.com/sigstore/cosign/blob/v3.1.3/doc/cosign_verify-blob.md),
[GoReleaser artifact 형식](https://goreleaser.com/customization/general/artifacts/).

## 로컬 entrypoint와 broker

`.make/build.mk`의 선언된 entrypoint를 사용한다.

| 대상                                    | 입력과 결과                                                       |
| --------------------------------------- | ----------------------------------------------------------------- |
| `release-check` / `release-healthcheck` | 고정 도구와 GoReleaser 설정 검사                                  |
| `release-snapshot` / `release-dry-run`  | 서명·게시 없는 동일 preview 경로                                  |
| `release-artifact-test`                 | tamper·잘못된 identity·기존 release·부분 실패 등 offline 회귀     |
| `release-artifact-prepare`              | hosted tag build의 RELEASE_TAG·RELEASE_SHA로 stage 생성           |
| `release-artifact-verify`               | RELEASE_ARTIFACT_DIR의 서명과 승인된 RELEASE_TAG·RELEASE_SHA 검증 |
| `release-artifact-publish`              | 같은 입력을 재검증한 뒤 broker credential로 원격 게시             |

작업 디렉터리는 승인된 source SHA를 checkout한 clean gzh-cli 작업 트리다.
`RELEASE_ARTIFACT_DIR`는 다운로드한 stage를 가리키며 기본값은 해당 트리의
`dist/publish`다. `RELEASE_TAG`·`RELEASE_SHA`는 명시 입력이며 프로그램이 새 버전을
선택하지 않는다. operator handoff는 실제 경로·source·환경이 확정된 뒤 별도 준비한다.

infra-ops `common/github/release-tokens.json`의 `main-release` group은 main source와 tap에
서로 다른 singleton Contents-write token을 주입한다. 환경 변수는 `GITHUB_TOKEN`과
`HOMEBREW_TAP_TOKEN`이다. 토큰 생성·범위/TTL 확인·회수는 infra-ops 소유 broker가
담당한다. 이 등록만으로 live App 설치·접근 권한이 검증됐다고 주장하지 않는다.
프로필·개인키·token 값을 소스나 로그에 남기지 않는다.

## 게시 순서와 실패

1. 정확한 파일 inventory·hash·서명, clean HEAD와 원격 tag의 source를 확인한다.
1. 공개/비공개 draft를 포함해 같은 tag의 기존 Release가 없음을 증명한다.
1. 기존 tap file의 SHA 또는 정규 404 부재를 확보한다.
1. 새 draft를 만들고 검증한 동일 asset bytes를 업로드해 server digest까지 확인한다.
1. 모든 asset이 확인된 GitHub Release를 공개한다.
1. 이미 공개된 asset URL을 가리키는 cask를 tap에 마지막으로 쓴다.

기존 cask는 old SHA로 CAS하고, 최초 cask는 SHA 없이 create한다. concurrent create나
내용 충돌을 overwrite로 해소하지 않는다. Release 공개가 실패하면 tap은 바뀌지 않는다.
tap이 실패하면 유효한 공개 Release와 이전 tap 상태를 보존하고 실패를 명확히 반환한다.

부분 실패에서는 어느 단계까지 수행됐는지와 불확실성을 보존한다. 전체 publisher를
자동 재실행하거나 기존 Release를 replace/delete하지 않는다. 실제 원격 상태를 확인한
후 누락 단계만 별도로 검토하는 복구가 필요하다.

## 도구 pin과 검증 경계

기존 Go1.27.3·GoReleaser v2.19.0·syft v1.51.1·cosign v3.1.3을 유지한다.
도구는 로컬 `bin/tools` 또는 SHA로 고정된 공식 Actions를 통해 준비한다.
`.make/tools.mk`와 release workflow의 세 version 선언은
`verify-release-pins`로 대조한다. cosign validator도 감사한 v3.1.3을 요구하므로 향후 pin
변경에는 verifier·회귀 검토가 함께 필요하다. 전역 도구 설정을 바꾸지 않는다.

제품의 핵심 로직 coverage 80%, 기동 50ms 미만·대다수 명령 100ms 미만과 전체 라이브러리
검증은 별도 출시 조건이다. unsigned snapshot이나 offline fake 테스트는 실제 OIDC 서명·
공개 Rekor 기록·원격 token 권한·게시·설치 성공을 대신하지 않는다.

## Versioning

The project follows [Semantic Versioning (SemVer)](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., `v1.2.3`)
- **Pre-releases**: `v1.2.3-rc.1`, `v1.2.3-beta.1`, `v1.2.3-alpha.1`

### Version Examples

| Version          | Type              | Description                          |
| ---------------- | ----------------- | ------------------------------------ |
| `v1.0.0`         | Major             | Breaking changes, new major features |
| `v1.1.0`         | Minor             | New features, backward compatible    |
| `v1.1.1`         | Patch             | Bug fixes, security updates          |
| `v1.2.0-rc.1`    | Release Candidate | Pre-release testing                  |
| `v1.2.0-beta.1`  | Beta              | Feature complete, testing            |
| `v1.2.0-alpha.1` | Alpha             | Early development, unstable          |
