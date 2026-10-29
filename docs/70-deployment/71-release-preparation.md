# gzh-cli 출시 준비 체크리스트

현재 첫 공개 후보는 **v0.2.0**이며 아직 게시되지 않았습니다.
이 체크리스트의 빈 항목은 실제 증거와 최종 승인으로만 완료합니다.
과거의 v1.0.0 준비 초안은 실제 출시나 품질 통과의 기록이 아닙니다.

## 후보·의존성

- [ ] 정확한 source commit·tree·tag 후보와 변경 범위를 동결한다.
- [ ] go.mod의 제품군 의존성이 채택된 안정 태그이며 replace·local workspace 우회가 없음을 확인한다.
- [ ] devbox의 local-dev 해제 절차를 수행하고 최종 source bytes를 다시 확인한다.
- [ ] CHANGELOG의 후보 절에 실제 게시 날짜를 넣고 [출시 노트](72-release-notes.md)와 맞춘다.

## 제품 게이트

- [ ] PRODUCT의 build·lint·test 기준을 소유 entrypoint로 통과한다.
- [ ] 명시적으로 채택된 core manifest와 원본 profile로 핵심 로직 coverage >=80%를 검증한다.
- [ ] 채택된 모집단·플랫폼·endpoint·통계로 기동 \<50ms, 대다수 명령 응답 \<100ms를 검증한다.
- [ ] package benchmark와 실제 CLI 프로세스 측정을 구분한다.
- [ ] 실제 명령·플래그·미지원 동작과 사용자 문서가 일치한다.
- [ ] 구현과 별개의 독립 리뷰가 exact source의 차단 결함 없이 통과한다.

## 비게시 산출물

- [ ] [출시 절차](70-releases.md)의 고정 release 도구 검사와 비게시 preview를 통과한다.
- [ ] 모든 지원 archive의 실행파일·LICENSE·문서·completion·manpage payload를 확인한다.
- [ ] archive·SBOM·cask·notes·manifest와 최종 checksum의 정확한 inventory·hash를 대조한다.
- [ ] snapshot을 정식 릴리스로 게시할 수 없음을 확인한다.
- [ ] 후보의 버전 출력과 플랫폼 buildinfo를 확인한다.

## 승인된 실제 출시

- [ ] 정확한 commit·tag·게시 내용과 실행 범위의 최종 승인을 확인한다.
- [ ] 기존 공개 tag나 release를 교체하지 않고 새 tag의 hosted 서명 빌드를 수행한다.
- [ ] 최종 checksum의 서명과 workflow·issuer·tag·SHA를 검증한다.
- [ ] infra-ops가 소유하는 consumer 쌍의 실제 저장소 권한과 별도 토큰 수명·회수 경계를 확인한다.
- [ ] 검증된 기존 bytes로 새 draft·asset 업로드·공개 release·tap CAS 순서를 완료한다.
- [ ] 부분 실패가 있으면 공개 상태를 먼저 확인하고 자동 재시도나 삭제로 숨기지 않는다.
- [ ] 실제 공개 URL·checksum·online audit·격리 fresh install·버전 출력을 검증한다.
- [ ] 첫 출시에는 이전 공개 버전이 없으므로 upgrade를 N/A로 기록하고, 사용자 설치본은 교체하지 않는다.
- [ ] 소유 작업 자원을 회수하고 결과를 단일 실행 기록에 누적한다.
