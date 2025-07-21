> ⚠️ 이 QA는 자동으로 검증할 수 없습니다.  
> 아래 절차에 따라 수동으로 확인해야 합니다.

### ✅ 수동 테스트 지침
- [ ] 실제 환경에서 테스트 수행
- [ ] 외부 서비스 연동 확인
- [ ] 사용자 시나리오 검증
- [ ] 결과 문서화

---

# title: GitHub 조직 관리 기능 QA 시나리오

## related_tasks
- /tasks/done/github-org-management__DONE_20250710.md
- /tasks/done/high-priority__DONE_20250711.md (GitHub 조직 관리 확장 부분)

## purpose  
GitHub 조직 관리 확장 기능들이 실제 GitHub API와 연동하여 정상 작동하는지 검증

## scenario

### 1. 리포지토리 설정 비교 도구 (`gz repo-config diff`)
1. 테스트 GitHub 조직 준비 (최소 3-5개 리포지토리)
2. 목표 설정 파일 (YAML) 작성
   - 브랜치 보호 규칙
   - 이슈/PR 템플릿 설정
   - Actions 권한 정책
3. `gz repo-config diff` 명령어 실행
4. 시각적 diff 출력 확인
5. 변경사항 요약 리포트 검증
6. 영향받는 리포지토리 목록 정확성 확인

### 2. 정책 준수 감사 고도화 (`gz repo-config audit`)
1. 정책 위반이 있는 리포지토리들 의도적 설정
2. `gz repo-config audit` 실행
3. HTML 리포트 생성 확인
4. 대시보드 형식 시각화 검증
5. 정책 위반 트렌드 분석 결과 확인
6. 규정 준수 점수 산출 알고리즘 검증
7. 자동 수정 제안 시스템 테스트
8. 위험도 평가 스코어링 검증

### 3. 웹훅 설정 관리
1. 테스트 웹훅 엔드포인트 설정
2. 조직 전체 웹훅 일괄 설정 테스트
3. GitHub 이벤트 수신 및 파싱 검증
   - push 이벤트
   - pull_request 이벤트  
   - issue 이벤트
4. 자동화 규칙 조건 평가 엔진 테스트
5. 자동화 액션 실행기 검증
6. 웹훅 상태 모니터링 대시보드 확인

### 4. GitHub Actions 권한 정책 관리
1. Actions 권한 정책 스키마 적용 테스트
2. 워크플로우 권한 감사 기능 검증
3. 정책 위반 감지 및 알림 테스트

### 5. 의존성 관리 정책
1. Dependabot 설정 관리 테스트
2. 보안 업데이트 정책 적용 검증
3. 의존성 버전 정책 강제 테스트

## expected_result
- **diff 도구**: 정확한 설정 차이점 시각화, 변경 영향도 분석
- **audit 도구**: 정책 위반 정확 감지, 위험도 기반 우선순위 제공
- **웹훅 관리**: 이벤트 실시간 수신, 자동화 규칙 정상 실행
- **권한 정책**: Actions 권한 정확 제어, 위반 사항 감지
- **의존성 정책**: Dependabot 설정 적용, 보안 업데이트 자동화

## test_data_requirements
- 테스트용 GitHub 조직 (3-5개 리포지토리)
- 다양한 설정 상태의 리포지토리들
- 정책 템플릿 파일들
- 웹훅 수신 테스트 서버

## tags
[qa], [integration], [manual], [github-api], [external-dependency]
