> ⚠️ 이 QA는 자동으로 검증할 수 없습니다.  
> 아래 절차에 따라 수동으로 확인해야 합니다.

### ✅ 수동 테스트 지침

- [ ] 실제 환경에서 테스트 수행
- [ ] 외부 서비스 연동 확인
- [ ] 사용자 시나리오 검증
- [ ] 결과 문서화
- [ ] 크로스 플랫폼 호환성 확인 (Linux/macOS/Windows)
- [ ] 네트워크 환경별 테스트 (WiFi/유선/VPN)
- [ ] 실제 GitHub/GitLab 조직 연동 테스트

---

# Manual QA Tests Summary

## Files Moved to Manual Testing:

1. **github-organization-management.qa.md** - ALL tests require GitHub org setup
2. **Network Environment Manual Tests** - Docker, K8s, VPN setup required
3. **UI/UX Manual Verification** - Visual inspection required

## How to Use:

1. Each manual test file contains agent-friendly command blocks
2. Copy the entire command block and paste into an agent session
3. Replace placeholder values (tokens, org names, etc.)
4. Run the commands and verify outputs

## Prerequisites for Manual Testing:

- GitHub organization with admin access
- GitHub personal access token with full repo permissions
- Docker running locally (for Docker tests)
- Kubernetes cluster access (for K8s tests)
- VPN configurations (for VPN tests)
- Cloud provider credentials (AWS/GCP/Azure)
