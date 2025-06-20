# 모든 리포지터리 동기화
gzh sync --all

# 특정 리포지터리만 동기화
gzh sync --repo myrepo

# dry-run (실제 push/pull 없이 변경사항만 확인)
gzh sync --all --dry-run

# 상태 확인
gzh status --all
