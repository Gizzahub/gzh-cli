# Synclone Test Data - Sample Repositories

이 문서는 synclone 기능을 테스트하기 위한 샘플 리포지터리와 조직 목록을 정리한 것입니다.

## 1. 테스트 권장 공개 조직 (GitHub)

### 1.1 소형 조직 (5-20개 리포지터리)

빠른 테스트와 기본 기능 검증에 적합:

#### Development Tools & CLI

```bash
# CLI 도구 개발 조직
gz synclone github -o golangci --target ./test-small/golangci    # ~10개 리포지터리
gz synclone github -o spf13 --target ./test-small/spf13         # ~15개 리포지터리 (Cobra, Viper 제작자)
gz synclone github -o urfave --target ./test-small/urfave       # ~12개 리포지터리 (CLI 라이브러리)
```

#### Container & Cloud Native (Small)

```bash
gz synclone github -o containerd --target ./test-small/containerd  # ~8개 리포지터리
gz synclone github -o jaegertracing --target ./test-small/jaeger   # ~15개 리포지터리
gz synclone github -o fluent --target ./test-small/fluent          # ~12개 리포지터리
```

#### Language-Specific

```bash
gz synclone github -o golang --target ./test-small/golang          # ~20개 리포지터리 (Go 공식)
gz synclone github -o nodejs --target ./test-small/nodejs          # ~15개 리포지터리 (Node.js 공식)
gz synclone github -o python --target ./test-small/python          # ~18개 리포지터리 (Python 공식)
```

### 1.2 중형 조직 (20-100개 리포지터리)

중간 규모 테스트와 병렬 처리 검증에 적합:

#### Monitoring & Observability

```bash
gz synclone github -o prometheus --target ./test-medium/prometheus     # ~50개 리포지터리
gz synclone github -o grafana --target ./test-medium/grafana          # ~80개 리포지터리
gz synclone github -o open-telemetry --target ./test-medium/otel      # ~60개 리포지터리
```

#### Cloud Native Foundation

```bash
gz synclone github -o etcd-io --target ./test-medium/etcd             # ~25개 리포지터리
gz synclone github -o helm --target ./test-medium/helm                # ~30개 리포지터리
gz synclone github -o istio --target ./test-medium/istio              # ~40개 리포지터리
```

#### HashiCorp Tools

```bash
gz synclone github -o hashicorp --target ./test-medium/hashicorp       # ~70개 리포지터리
```

### 1.3 대형 조직 (100+ 리포지터리)

대규모 테스트와 최적화 기능 검증에 적합:

#### Cloud Native Computing Foundation

```bash
gz synclone github -o cncf --target ./test-large/cncf                  # ~200개 리포지터리
gz synclone github -o kubernetes --target ./test-large/kubernetes      # ~150개 리포지터리
gz synclone github -o kubernetes-sigs --target ./test-large/k8s-sigs   # ~300개 리포지터리
```

#### Major Tech Companies

```bash
gz synclone github -o microsoft --target ./test-large/microsoft        # ~500개+ 리포지터리 (매우 대규모)
gz synclone github -o google --target ./test-large/google              # ~400개+ 리포지터리
gz synclone github -o facebook --target ./test-large/facebook          # ~200개+ 리포지터리
```

#### Open Source Foundations

```bash
gz synclone github -o apache --target ./test-large/apache              # ~1000개+ 리포지터리 (초대규모)
gz synclone github -o eclipse --target ./test-large/eclipse            # ~800개+ 리포지터리
```

## 2. 특수 목적 테스트 조직

### 2.1 언어별 테스트

#### Go Language Ecosystem

```bash
gz synclone github -o golang --language Go                             # Go 공식 리포지터리
gz synclone github -o hashicorp --language Go                          # Go 기반 인프라 도구
gz synclone github -o kubernetes --language Go                         # Go 기반 클라우드 네이티브
gz synclone github -o prometheus --language Go                         # Go 기반 모니터링
```

#### JavaScript/TypeScript Ecosystem

```bash
gz synclone github -o nodejs --language JavaScript                     # Node.js 공식
gz synclone github -o microsoft --language TypeScript                  # TypeScript 관련
gz synclone github -o facebook --language JavaScript                   # React 등
```

#### Python Ecosystem

```bash
gz synclone github -o python --language Python                         # Python 공식
gz synclone github -o pallets --language Python                        # Flask, Jinja2 등
gz synclone github -o psf --language Python                            # Python Software Foundation
```

### 2.2 프로젝트 특성별 테스트

#### CLI Tools & Developer Tools

```bash
gz synclone github -o cli --topics "cli,command-line,developer-tools"
gz synclone github -o github --topics "cli,git,github"
```

#### Container & Kubernetes

```bash
gz synclone github -o kubernetes --topics "kubernetes,container,docker"
gz synclone github -o docker --topics "docker,container"
```

#### Monitoring & Logging

```bash
gz synclone github -o prometheus --topics "monitoring,metrics"
gz synclone github -o elastic --topics "logging,search,elasticsearch"
```

### 2.3 크기별 테스트 데이터

#### 작은 리포지터리 (< 1MB)

```bash
gz synclone github -o awesome-lists --size-limit 1024                  # 문서 위주
gz synclone github -o sindresorhus --size-limit 1024                   # 작은 유틸리티
```

#### 중간 리포지터리 (1-10MB)

```bash
gz synclone github -o golang --size-limit 10240                        # 일반적인 프로젝트
gz synclone github -o prometheus --size-limit 10240
```

#### 큰 리포지터리 (10MB+)

```bash
gz synclone github -o kubernetes --min-size 10240                      # 대규모 프로젝트
gz synclone github -o tensorflow --min-size 10240                      # ML 프로젝트
```

## 3. 인기도별 테스트 데이터

### 3.1 스타 수 기준

#### 초보자 프로젝트 (< 100 stars)

```bash
gz synclone github -o your-username --max-stars 100                    # 개인 프로젝트
```

#### 인기 프로젝트 (100-1000 stars)

```bash
gz synclone github -o prometheus --min-stars 100 --max-stars 1000     # 중간 인기도
gz synclone github -o grafana --min-stars 100 --max-stars 1000
```

#### 매우 인기 프로젝트 (1000+ stars)

```bash
gz synclone github -o kubernetes --min-stars 1000                     # 고인기도
gz synclone github -o microsoft --min-stars 1000
```

### 3.2 활발도 기준

#### 최근 활발한 프로젝트

```bash
gz synclone github -o kubernetes --updated-after 2024-01-01           # 최근 1년
gz synclone github -o prometheus --updated-after 2024-06-01           # 최근 6개월
gz synclone github -o grafana --updated-after 2024-09-01              # 최근 3개월
```

#### 오래된 프로젝트

```bash
gz synclone github -o apache --updated-before 2023-01-01              # 1년 이상 미업데이트
```

## 4. 실제 테스트 시나리오별 데이터 세트

### 4.1 성능 테스트용 데이터 세트

#### 빠른 성능 테스트 (< 5분)

```bash
# 소규모 조직들로 빠른 테스트
gz synclone github -o golangci --target ./perf-test-small
gz synclone github -o spf13 --target ./perf-test-small2
gz synclone github -o urfave --target ./perf-test-small3
```

#### 중간 성능 테스트 (5-15분)

```bash
# 중규모 조직으로 병렬 처리 테스트
gz synclone github -o prometheus --target ./perf-test-medium --parallel 10
gz synclone github -o grafana --target ./perf-test-medium2 --parallel 10
```

#### 대규모 성능 테스트 (15분+)

```bash
# 대규모 조직으로 최적화 테스트
gz synclone github -o kubernetes --target ./perf-test-large --optimized --parallel 20
gz synclone github -o cncf --target ./perf-test-large2 --streaming --memory-limit 1GB
```

### 4.2 필터링 기능 테스트용 데이터

#### 패턴 매칭 테스트

```bash
# kubectl 관련만
gz synclone github -o kubernetes --include "^kubectl.*" --target ./filter-test-kubectl

# prometheus 관련만
gz synclone github -o prometheus --include "^prometheus.*" --target ./filter-test-prom

# 아카이브 제외
gz synclone github -o kubernetes --exclude ".*-archive$|.*-deprecated$" --target ./filter-test-active
```

#### 언어 필터링 테스트

```bash
# Go 전용 테스트
gz synclone github -o hashicorp --language Go --target ./filter-test-go
gz synclone github -o kubernetes --language Go --target ./filter-test-go-k8s

# JavaScript 전용 테스트
gz synclone github -o facebook --language JavaScript --target ./filter-test-js
gz synclone github -o nodejs --language JavaScript --target ./filter-test-js-node
```

### 4.3 에러 처리 테스트용 데이터

#### 존재하지 않는 조직 (404 에러 테스트)

```bash
gz synclone github -o nonexistent-org-12345 --target ./error-test-404
```

#### 권한 없는 프라이빗 조직 (403 에러 테스트)

```bash
# 프라이빗 조직명 (접근 권한 없음)
gz synclone github -o super-secret-private-org --include-private --target ./error-test-403
```

#### 빈 조직 (빈 결과 테스트)

```bash
# 리포지터리가 매우 적은 새 조직이나 개인 계정
gz synclone github -o new-empty-user --target ./error-test-empty
```

## 5. GitLab 테스트 데이터

### 5.1 공개 GitLab 그룹

#### GitLab 공식

```bash
gz synclone gitlab -g gitlab-org --target ./gitlab-test/gitlab-org      # GitLab 자체
gz synclone gitlab -g gitlab-com --target ./gitlab-test/gitlab-com      # GitLab.com 관련
```

#### GNOME 프로젝트

```bash
gz synclone gitlab -g GNOME --target ./gitlab-test/gnome               # GNOME 데스크톱
```

#### KDE 프로젝트

```bash
gz synclone gitlab -g kde --target ./gitlab-test/kde                   # KDE 데스크톱
```

### 5.2 하위 그룹 테스트

```bash
# 재귀적 하위 그룹 클로닝
gz synclone gitlab -g gitlab-org --recursive --target ./gitlab-test/recursive
```

## 6. Gitea 테스트 데이터

### 6.1 공개 Gitea 인스턴스

#### Gitea 공식 (codeberg.org)

```bash
gz synclone gitea -o gitea --api-url https://codeberg.org --target ./gitea-test/gitea
```

#### 기타 공개 Gitea 인스턴스

```bash
# 예시: 대학이나 오픈소스 프로젝트의 Gitea 인스턴스
gz synclone gitea -o example-org --api-url https://git.example.org --target ./gitea-test/example
```

## 7. 테스트 실행 가이드

### 7.1 단계별 테스트 실행

#### 1단계: 기본 기능 테스트

```bash
# 작은 조직으로 빠른 테스트
gz synclone github -o golangci --target ./test-stage-1
echo "예상 시간: 1-2분, 예상 리포지터리: ~10개"
```

#### 2단계: 필터링 기능 테스트

```bash
# 필터링 옵션 테스트
gz synclone github -o kubernetes --include "^kubectl.*" --target ./test-stage-2
echo "예상 시간: 2-3분, 예상 리포지터리: ~5개"
```

#### 3단계: 병렬 처리 테스트

```bash
# 중간 크기 조직으로 병렬 처리
gz synclone github -o prometheus --parallel 10 --target ./test-stage-3
echo "예상 시간: 3-5분, 예상 리포지터리: ~50개"
```

#### 4단계: 대규모 테스트

```bash
# 대규모 조직으로 최적화 기능 테스트
gz synclone github -o kubernetes --optimized --parallel 20 --target ./test-stage-4
echo "예상 시간: 10-15분, 예상 리포지터리: ~150개"
```

### 7.2 테스트 결과 확인 스크립트

```bash
#!/bin/bash

# 테스트 결과 검증 함수
verify_test_result() {
    local test_dir=$1
    local expected_min=$2
    local test_name=$3
    
    if [ -d "$test_dir" ]; then
        local repo_count=$(find "$test_dir" -name ".git" -type d | wc -l)
        local gzh_files=$(find "$test_dir" -name "gzh.yaml" | wc -l)
        
        echo "=== $test_name ==="
        echo "디렉토리: $test_dir"
        echo "리포지터리 수: $repo_count"
        echo "gzh.yaml 파일 수: $gzh_files"
        
        if [ "$repo_count" -ge "$expected_min" ]; then
            echo "✅ 성공: 예상 최소 개수($expected_min) 이상"
        else
            echo "❌ 실패: 예상보다 적음 ($repo_count < $expected_min)"
        fi
        
        if [ "$gzh_files" -eq 1 ]; then
            echo "✅ gzh.yaml 파일 정상 생성"
        else
            echo "❌ gzh.yaml 파일 누락 또는 중복"
        fi
        echo ""
    else
        echo "❌ 테스트 디렉토리 없음: $test_dir"
        echo ""
    fi
}

# 테스트 결과 검증 실행
verify_test_result "./test-stage-1" 5 "1단계: 기본 기능 테스트"
verify_test_result "./test-stage-2" 3 "2단계: 필터링 기능 테스트" 
verify_test_result "./test-stage-3" 30 "3단계: 병렬 처리 테스트"
verify_test_result "./test-stage-4" 100 "4단계: 대규모 테스트"
```

## 8. 주의사항 및 권장사항

### 8.1 테스트 환경 준비

1. **네트워크**: 안정적인 인터넷 연결 필요
1. **디스크 공간**: 대규모 테스트시 10GB+ 여유 공간 권장
1. **GitHub API 제한**: 토큰 없이는 시간당 60회 제한
1. **메모리**: 대규모 병렬 처리시 충분한 RAM 필요

### 8.2 테스트 순서 권장사항

1. **소규모** → **중규모** → **대규모** 순서로 테스트
1. **기본 기능** → **고급 기능** → **최적화 기능** 순서
1. **성공 케이스** → **에러 케이스** → **엣지 케이스** 순서

### 8.3 정리 스크립트

```bash
#!/bin/bash
# 테스트 후 정리 스크립트

echo "테스트 디렉토리 정리 중..."

# 모든 테스트 디렉토리 삭제
rm -rf ./test-* ./perf-test-* ./filter-test-* ./error-test-* ./gitlab-test ./gitea-test

# 상태 파일 정리 (선택적)
# gz synclone state clean --age 1d

echo "정리 완료!"
```

이 테스트 데이터를 활용하면 synclone의 모든 기능을 체계적이고 안전하게 검증할 수 있습니다.
