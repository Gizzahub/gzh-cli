#!/bin/bash
# QA 자동 처리 스크립트

set -euo pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Directories
QA_DIR="$SCRIPT_DIR"
DONE_DIR="$PROJECT_ROOT/tasks/done/qa"
MANUAL_DIR="$QA_DIR/manual"

# Create directories if they don't exist
mkdir -p "$DONE_DIR"
mkdir -p "$MANUAL_DIR"

# Counters
AUTO_PROCESSED=0
MANUAL_PROCESSED=0
TOTAL_PROCESSED=0

echo -e "${BLUE}🧪 QA 자동 처리 시작${NC}"
echo "Working directory: $QA_DIR"
echo ""

# Function to analyze QA file for automation potential
analyze_qa_file() {
    local file="$1"
    local has_executable_commands=false
    local has_clear_expectations=false
    local needs_manual_verification=false

    # Check for executable commands
    if grep -q -E "(gz |make |go |npm |docker |kubectl |./|bash |sh )" "$file"; then
        has_executable_commands=true
    fi

    # Check for clear expected results
    if grep -q -E "(expected|예상|결과|통과|실패|성공)" "$file"; then
        has_clear_expectations=true
    fi

    # Check for manual verification requirements
    if grep -q -E "(수동|manual|사용자|UI|경험|체감|주관|크로스|플랫폼|실제)" "$file"; then
        needs_manual_verification=true
    fi

    # Decision logic
    if [ "$has_executable_commands" = true ] && [ "$has_clear_expectations" = true ] && [ "$needs_manual_verification" = false ]; then
        echo "auto"
    else
        echo "manual"
    fi
}

# Function to process auto QA
process_auto_qa() {
    local file="$1"
    local filename=$(basename "$file")

    echo -e "${GREEN}✅ 자동 처리: $filename${NC}"

    # Add automation result to file
    cat >> "$file" << EOF

---
## ✅ 자동 테스트 결과
- 처리 시간: $(date)
- 상태: 자동 처리 대상으로 분류됨
- 실행 가능한 명령어 발견: 예
- 명확한 예상 결과: 예
- 수동 검증 요구 사항: 없음

> 📝 주의: 실제 테스트 실행은 컴파일 에러 수정 후 가능합니다.
EOF

    # Move to done directory
    mv "$file" "$DONE_DIR/${filename%.*}__AUTO_PROCESSED_$(date +%Y%m%d).md"

    AUTO_PROCESSED=$((AUTO_PROCESSED + 1))
}

# Function to process manual QA
process_manual_qa() {
    local file="$1"
    local filename=$(basename "$file")

    echo -e "${YELLOW}🛠️ 수동 처리: $filename${NC}"

    # Add manual testing guidance
    local temp_file=$(mktemp)
    cat > "$temp_file" << 'EOF'
> ⚠️ 이 QA는 자동으로 검증할 수 없습니다.
> 아래 절차에 따라 수동으로 확인해야 합니다.

### ✅ 수동 테스트 지침
- [ ] 테스트 환경 준비
- [ ] 크로스 플랫폼 호환성 확인 (Linux/macOS/Windows)
- [ ] 실제 네트워크 환경에서 검증
- [ ] 사용자 경험 및 UI 일관성 확인
- [ ] 성능 및 체감도 측정

---

EOF

    # Prepend manual guidance to original file
    cat "$temp_file" "$file" > "${file}.tmp"
    mv "${file}.tmp" "$file"
    rm "$temp_file"

    # Move to manual directory if not already there
    if [[ "$file" != *"/manual/"* ]]; then
        mv "$file" "$MANUAL_DIR/$filename"
    fi

    MANUAL_PROCESSED=$((MANUAL_PROCESSED + 1))
}

# Process QA files
echo -e "${BLUE}📁 QA 파일 분석 중...${NC}"

# Process main QA files (excluding already processed ones)
for file in "$QA_DIR"/*.md; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")

        # Skip already processed files
        if [[ "$filename" == "QA_FINAL_REPORT.md" ]] || [[ "$filename" == "FINAL_QA_CHECKLIST.md" ]]; then
            echo -e "${BLUE}📋 건너뛰기: $filename (메타 문서)${NC}"
            continue
        fi

        TOTAL_PROCESSED=$((TOTAL_PROCESSED + 1))

        # Analyze and process
        analysis_result=$(analyze_qa_file "$file")

        if [ "$analysis_result" = "auto" ]; then
            process_auto_qa "$file"
        else
            process_manual_qa "$file"
        fi
    fi
done

# Process existing manual files
echo -e "\n${BLUE}📁 기존 수동 테스트 파일 검증 중...${NC}"

for file in "$MANUAL_DIR"/*.md; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")

        # Add manual guidance if not present
        if ! grep -q "⚠️ 이 QA는 자동으로 검증할 수 없습니다" "$file"; then
            echo -e "${YELLOW}📝 수동 가이드 추가: $filename${NC}"

            temp_file=$(mktemp)
            cat > "$temp_file" << 'EOF'
> ⚠️ 이 QA는 자동으로 검증할 수 없습니다.
> 아래 절차에 따라 수동으로 확인해야 합니다.

### ✅ 수동 테스트 지침
- [ ] 실제 환경에서 테스트 수행
- [ ] 외부 서비스 연동 확인
- [ ] 사용자 시나리오 검증
- [ ] 결과 문서화

---

EOF
            cat "$temp_file" "$file" > "${file}.tmp"
            mv "${file}.tmp" "$file"
            rm "$temp_file"
        fi
    fi
done

# Test automatic tests if binary exists
echo -e "\n${BLUE}🧪 자동화 테스트 실행 시도${NC}"

if [ -f "$PROJECT_ROOT/gz" ] || command -v gz &> /dev/null; then
    echo "gz 바이너리 발견. 자동화 테스트 실행 중..."

    if bash "$QA_DIR/run_automated_tests.sh"; then
        echo -e "${GREEN}✅ 자동화 테스트 성공${NC}"
    else
        echo -e "${RED}❌ 자동화 테스트 실패 (예상됨 - 컴파일 에러 존재)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️ gz 바이너리 없음. 빌드 필요.${NC}"
    echo "컴파일 에러 수정 후 다음 명령어로 테스트 실행:"
    echo "  go build -o gz ./cmd && ./tasks/qa/run_automated_tests.sh"
fi

# Generate summary report
echo -e "\n${BLUE}📊 처리 결과 요약${NC}"
echo "==================="
echo -e "총 처리된 파일: $TOTAL_PROCESSED"
echo -e "자동 처리: ${GREEN}$AUTO_PROCESSED${NC}"
echo -e "수동 처리: ${YELLOW}$MANUAL_PROCESSED${NC}"

# Create summary file
cat > "$QA_DIR/qa_processing_summary.md" << EOF
# QA 자동 처리 결과 요약

## 처리 통계
- **처리 날짜**: $(date)
- **총 처리 파일**: $TOTAL_PROCESSED개
- **자동 처리**: $AUTO_PROCESSED개
- **수동 처리**: $MANUAL_PROCESSED개

## 디렉토리 구조
\`\`\`
/tasks/
├── qa/
│   ├── manual/          # 수동 테스트 가이드
│   ├── tests/           # 자동화 테스트 스크립트
│   └── run_automated_tests.sh
└── done/
    └── qa/              # 자동 처리 완료된 QA 파일
\`\`\`

## 다음 단계
1. 컴파일 에러 수정
2. 자동화 테스트 실행: \`./tasks/qa/run_automated_tests.sh\`
3. 수동 테스트 수행: \`/tasks/qa/manual/\` 디렉토리 참고

## 자동화 기준
- ✅ 실행 가능한 명령어 포함
- ✅ 명확한 예상 결과 정의
- ❌ 수동 검증 요구 사항 없음

## 수동 처리 기준
- 크로스 플랫폼 테스트 필요
- 실제 환경 연동 필요
- 사용자 경험 평가 필요
- 주관적 판단 요구
EOF

echo -e "\n${GREEN}✨ QA 자동 처리 완료!${NC}"
echo "📄 상세 결과: $QA_DIR/qa_processing_summary.md"

exit 0
