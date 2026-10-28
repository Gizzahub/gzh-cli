# 진단·저장소 목록의 평가 상태

`gz doctor metrics`는 기본 threshold75를 사용한다. table·JSON·YAML·only-issues 모두
보고서 저장과 출력을 성공시킨 뒤 같은 `score < threshold` 판정을 적용한다.
기준과 같은 점수는 통과한다. 저장·출력 오류는 threshold 오류보다 먼저 반환한다.
기준 미달이어도 성공적으로 생성한 report payload는 보존한다.

현재 이력 backend가 없으므로 `history_assessment`는 `not_assessed`다.
기존 `trends` 필드의 빈 문자열·0은 이력이 측정됐거나 안정적이라는 뜻이 아니다.
기본 실행은 유지하지만 `--historical`을 명시하면 값7을 포함해 분석·외부 도구 실행 전에
미지원 오류를 반환한다. 실제 이력 수집이나 추세 알고리즘을 제공한다는 뜻이 아니다.

`gz repo-config list`의 `compliance_assessment`도 `not_assessed`다.
현재 목록 조회는 실제 repository audit를 수행하지 않는다. table은 compliance와 issues를
`NOT_ASSESSED`로 표시하고 준수율을 계산하지 않는다. 빈 목록에서도 같은 의미를 유지한다.

JSON/YAML의 기존 `compliant` bool과 `issues` int는 타입 호환을 위해 유지한다.
`compliant`의 기존 template 할당 판정이나 `issues: 0`을 실제 준수 판정·문제0건의
관측으로 사용하면 안 된다. 소비자는 먼저 `compliance_assessment`를 확인해야 한다.
기존 field만 읽는 소비자의 오해 가능성이 남으므로 새 상태를 읽도록 갱신해야 한다.
새 field를 허용하지 않는 엄격한 schema 소비자도 schema 갱신이 필요하다.

예시:

```sh
gz doctor metrics --format json --threshold 80
gz doctor metrics --only-issues --threshold 80
gz repo-config list --org myorg --format json
```

`repo-config apply`는 기존 명시적 미지원 계약을 유지한다. 이 변경은 실제 감사·적용
backend를 구현하거나 클라우드 인증·사용자 설정을 바꾸지 않는다.
