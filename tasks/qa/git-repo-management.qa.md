# ✅ Git Repository Management QA 시나리오

- related_tasks:
  - `/tasks/done/05-git-repo-command-structure__DONE_20250728.md`
  - `/tasks/done/06-git-repo-provider-abstraction__DONE_20250728.md`
  - `/tasks/done/07-git-repo-clone-implementation__DONE_20250128.md`
  - `/tasks/done/08-git-repo-lifecycle-commands__DONE_20250128.md`
  - `/tasks/done/09-git-repo-sync-implementation__DONE_20250728.md`
- purpose: gz git repo 명령어 패밀리의 모든 기능이 정상 작동하는지 통합 검증
- tags: [qa, e2e, manual, grouped]

---

## 🧪 테스트 시나리오

### 1. Repository 조회 및 클론 기능 검증
1. 조직별 저장소 목록 조회
   ```bash
   gz git repo list --provider github --org gizzahub
   gz git repo list --provider gitlab --group mygroup
   ```
2. 필터링 기능 테스트
   ```bash
   gz git repo list --provider github --org gizzahub --language Go --visibility private
   ```
3. 저장소 클론 실행
   ```bash
   gz git repo clone --provider github --org gizzahub --target ./repos --match "api-*"
   ```

### 2. Repository 생명주기 관리 검증
1. 새 저장소 생성
   ```bash
   gz git repo create --provider github --org test-org --name qa-test-repo --private --description "QA test repository"
   ```
2. 저장소 정보 조회
   ```bash
   gz git repo get --provider github --org test-org --repo qa-test-repo
   ```
3. 저장소 아카이브
   ```bash
   gz git repo archive --provider github --org test-org --repo qa-test-repo
   ```
4. 저장소 삭제
   ```bash
   gz git repo delete --provider github --org test-org --repo qa-test-repo --confirm
   ```

### 3. Provider 추상화 계층 검증
1. 동일한 명령어로 다른 provider 테스트
   ```bash
   gz git repo list --provider github --org myorg
   gz git repo list --provider gitlab --group mygroup
   gz git repo list --provider gitea --org myorg
   ```
2. Provider별 고유 기능 및 제약사항 확인
3. 에러 처리 및 메시지 일관성 검증

### 4. Cross-Provider 동기화 검증
1. 단일 저장소 동기화
   ```bash
   gz git repo sync --from github:org/repo --to gitlab:group/repo --create-missing
   ```
2. 조직 전체 동기화
   ```bash
   gz git repo sync --from github:sourceorg --to gitlab:targetgroup --dry-run
   ```
3. 충돌 해결 및 에러 처리 확인
4. 동기화 진행률 및 로그 확인

### 5. 고급 기능 및 옵션 검증
1. 병렬 처리 (`--parallel 5`)
2. Dry-run 모드 (`--dry-run`)
3. 패턴 매칭 (`--match`, `--exclude`)
4. 출력 형식 (`--format table|json|yaml`)
5. 검색 기능 (`gz git repo search --query "golang api"`)

---

## ✅ 기대 결과

- 모든 provider에서 일관된 명령어 인터페이스 제공
- Repository CRUD 작업이 각 platform에서 정상 실행
- Cross-provider 동기화가 데이터 손실 없이 실행
- 에러 상황에서 명확한 메시지와 복구 방법 제시
- 대용량 작업시 안정적인 병렬 처리 및 진행률 표시
- 모든 명령어가 예상된 출력 형식으로 결과 반환

---

## 🚨 검증 포인트

- Provider간 API 차이점을 올바르게 추상화했는지
- 권한 에러, 네트워크 에러 등 다양한 실패 시나리오 처리
- 대용량 조직의 저장소 목록 조회 성능
- Cross-provider 동기화시 메타데이터 보존
- 동시 실행시 rate limiting 및 에러 복구
