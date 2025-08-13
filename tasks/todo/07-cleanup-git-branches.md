# TODO: Git 브랜치 정리

- status: [ ]
- priority: low (P3)
- category: project-maintenance  
- estimated_effort: 15분
- depends_on: []
- spec_reference: `git branch -av` 명령어 결과

## 📋 작업 개요

프로젝트 저장소에 존재하는 오래되거나 더 이상 필요하지 않은 브랜치들을 정리하여 저장소를 깔끔하게 유지합니다. 현재 여러 개의 오래된 기능 브랜치들이 남아있는 상태입니다.

## 🎯 정리 대상 브랜치들

### 1. **로컬 브랜치들**
- [ ] **remove-unused-packages-legacy** (c9d884b) 
  - 커밋: "refactor(sonnet): remove unused internal/api package"
  - 상태: 이미 완료된 리팩토링 작업으로 보임
  - 조치: 메인 브랜치에 머지되었는지 확인 후 삭제

- [ ] **simplify-container-usage** (f484583)
  - 커밋: "docs(sonnet): complete architecture simplification documentation"  
  - 상태: 문서화 작업 완료된 것으로 보임
  - 조치: 메인 브랜치에 머지되었는지 확인 후 삭제

### 2. **원격 브랜치들**
- [ ] **origin/add-claude-github-actions-1753076381793** (da78109)
  - 여러 중복된 GitHub Actions 관련 브랜치들
  - 임시 브랜치로 보이며 정리 필요

- [ ] **origin/add-claude-github-actions-1753076718544** (da78109)
  - 위와 동일한 커밋, 중복 브랜치

- [ ] **origin/add-claude-github-actions-1753079180841** (b541f33)
  - "Claude Code Review workflow" 커밋
  - GitHub Actions 워크플로우 관련

## 🔧 브랜치 정리 절차

### 1. 현재 상태 분석
```bash
# 모든 브랜치 상태 확인
git branch -av

# 각 브랜치가 메인 브랜치에 머지되었는지 확인
git branch --merged master
git branch --merged develop

# 각 브랜치의 마지막 커밋 확인
git for-each-ref --format='%(refname:short) %(committerdate) %(authorname)' --sort=-committerdate refs/heads/
```

### 2. 머지 상태 확인
```bash
# 특정 브랜치가 다른 브랜치에 포함되어 있는지 확인
git merge-base --is-ancestor remove-unused-packages-legacy develop
git merge-base --is-ancestor simplify-container-usage develop

# 브랜치 간 차이점 확인
git diff develop..remove-unused-packages-legacy
git diff develop..simplify-container-usage
```

### 3. 안전한 삭제 절차
```bash
# 로컬 브랜치 삭제 (머지 확인 후)
git branch -d remove-unused-packages-legacy
git branch -d simplify-container-usage

# 강제 삭제 (필요한 경우만)
git branch -D branch-name

# 원격 브랜치 삭제
git push origin --delete add-claude-github-actions-1753076381793
git push origin --delete add-claude-github-actions-1753076718544  
git push origin --delete add-claude-github-actions-1753079180841
```

## 📋 브랜치별 상세 분석

### remove-unused-packages-legacy
```
현재 상태: 로컬에만 존재
마지막 커밋: c9d884b "refactor(sonnet): remove unused internal/api package"
작업 내용: 사용하지 않는 패키지 제거
처리 방법: develop 브랜치와 비교 후 내용이 이미 포함되었으면 삭제
```

### simplify-container-usage  
```
현재 상태: 로컬에만 존재
마지막 커밋: f484583 "docs(sonnet): complete architecture simplification documentation"
작업 내용: 아키텍처 문서화 작업
처리 방법: 문서가 현재 버전에 반영되었는지 확인 후 결정
```

### GitHub Actions 관련 브랜치들
```
현재 상태: 원격에 존재, 유사한 이름의 여러 브랜치
작업 내용: GitHub Actions 워크플로우 설정
처리 방법: 현재 .github/workflows/ 디렉터리와 비교하여 이미 반영된 것들 삭제
```

## 🧪 정리 전 확인사항

### 1. 브랜치 내용 보존 확인
```bash
# 각 브랜치의 고유한 커밋들 확인
git rev-list --oneline develop..remove-unused-packages-legacy
git rev-list --oneline develop..simplify-container-usage

# 중요한 변경사항이 있는지 파일별 확인
git diff develop remove-unused-packages-legacy --name-only
git diff develop simplify-container-usage --name-only
```

### 2. 현재 워킹 디렉터리 확인
```bash  
# 현재 브랜치 확인
git branch --show-current

# 변경되지 않은 파일들 확인
git status
```

### 3. 원격 상태 동기화
```bash
# 원격 브랜치 정보 업데이트
git fetch --all --prune

# 더 이상 존재하지 않는 원격 브랜치 참조 정리
git remote prune origin
```

## ✅ 완료 기준

### 브랜치 정리 완료
- [ ] 오래된 로컬 기능 브랜치 2개 삭제 완료
- [ ] 임시 GitHub Actions 브랜치 3개 삭제 완료
- [ ] 정리 후 `git branch -av` 출력 결과 깔끔함
- [ ] 원격 저장소에서도 불필요한 브랜치 제거 완료

### 안전성 확인
- [ ] 삭제 전 중요한 변경사항 백업 또는 머지 완료
- [ ] 현재 작업 중인 브랜치에 영향 없음
- [ ] 팀원들이 사용 중인 브랜치 보존
- [ ] CI/CD 파이프라인에 영향 없음

### 저장소 상태 개선
- [ ] 브랜치 목록이 현재 개발 상황 반영
- [ ] 혼란을 줄 수 있는 오래된 브랜치 제거
- [ ] 새로운 기여자가 이해하기 쉬운 브랜치 구조

## 🚀 커밋 메시지 가이드

이 작업은 주로 브랜치 삭제 작업이므로 별도 커밋이 필요하지 않지만, 필요한 경우:

```
chore(claude-opus): 오래된 브랜치 정리

- remove-unused-packages-legacy 브랜치 삭제 (내용 이미 반영됨)
- simplify-container-usage 브랜치 삭제 (문서화 완료)
- GitHub Actions 관련 임시 브랜치 3개 정리
- 저장소 브랜치 구조 정리 및 가독성 향상

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 💡 구현 힌트

1. **안전 우선**: 삭제 전 반드시 내용 확인 및 백업
2. **단계별 처리**: 로컬 브랜치부터 정리 후 원격 브랜치 처리  
3. **팀 협의**: 다른 개발자가 사용 중일 수 있는 브랜치는 확인 후 삭제
4. **문서화**: 삭제한 브랜치의 내용이 어디에 반영되었는지 기록

## 🔗 관련 작업

이 작업은 다음과 연계됩니다:
- 전반적인 프로젝트 정리 작업
- Git 워크플로우 개선
- 새로운 기여자를 위한 저장소 정리

## ⚠️ 주의사항

### 삭제 전 주의점
- **팀원 확인**: 다른 개발자가 해당 브랜치에서 작업 중인지 확인
- **CI/CD 영향**: 파이프라인에서 참조하는 브랜치인지 확인
- **백업**: 중요한 변경사항이 있을 수 있으므로 삭제 전 내용 확인

### 삭제 시 주의점  
- **강제 삭제 금지**: 가능한 한 `-d` 옵션으로 안전하게 삭제
- **원격 우선**: 로컬 삭제 전에 원격 브랜치 상태 확인
- **현재 브랜치**: 삭제하려는 브랜치에 checkout 상태가 아닌지 확인

### 삭제 후 확인사항
- **git fetch 정리**: 더 이상 존재하지 않는 원격 브랜치 참조 정리
- **IDE 갱신**: 개발 환경에서 브랜치 목록 새로고침
- **문서 업데이트**: 필요시 개발 가이드 문서 업데이트

## 📊 정리 효과

### 긍정적 효과
- **가독성 향상**: 브랜치 목록이 현재 상황을 명확히 반영
- **혼란 감소**: 오래된 브랜치로 인한 개발자 혼란 방지
- **저장소 정리**: 클린한 저장소 상태로 새로운 기여자 친화적

### 유지보수 개선
- **브랜치 전략 명확화**: 활성 브랜치들만 남겨서 개발 흐름 파악 용이
- **성능 향상**: 브랜치 목록 조회 속도 개선
- **협업 효율성**: 팀원들이 집중해야 할 브랜치 명확화