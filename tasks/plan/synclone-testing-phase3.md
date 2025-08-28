# Phase 3: Docker 기반 격리 환경

## 📋 개요

**목표**: 완전히 격리된 환경에서 실제 Git 서버와 연동 테스트  
**소요 시간**: 7일  
**우선순위**: 중간 (CI/CD 통합 시 필요)  
**전제 조건**: Phase 1, 2 완료

## 🎯 구현 범위

- **기반**: 기존 `test/integration/docker/` 확장
- **목적**: 실제 운영 환경과 유사한 조건에서 synclone 동작 검증
- **격리**: 외부 의존성 없는 완전 자립적 테스트 환경

## 🏗️ 테스트 환경 아키텍처

### 컨테이너 구성
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Gitea Server  │    │  Redis Cache    │    │  Test Runner    │
│   (Git API)     │◄───┤  (State Store)  │◄───┤  (gz CLI)       │
│   Port: 3000    │    │   Port: 6379    │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         ▲                       ▲                       ▲
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  PostgreSQL DB  │
                    │  (Gitea Data)   │
                    │   Port: 5432    │
                    └─────────────────┘
```

## 📝 고급 테스트 케이스 (25개)

### 1. 실제 API 연동 (8개)
- **Git 서버 API 호출 검증**
  - 저장소 목록 조회 API
  - 저장소 상세 정보 API
  - 브랜치 목록 API
  - 태그 목록 API
- **인증 토큰 처리**
  - 유효한 토큰으로 인증 성공
  - 만료된 토큰 처리
  - 권한 없는 토큰 처리
- **Rate limiting 동작 확인**
  - API 호출 빈도 제한
  - Rate limit 초과 시 대응

### 2. 동시성 테스트 (8개)
- **여러 프로세스에서 동시 synclone 실행**
  - 2개 프로세스 동시 실행
  - 5개 프로세스 동시 실행
  - 10개 프로세스 동시 실행
- **락 파일 처리 검증**
  - 락 파일 생성/삭제
  - 데드락 방지
  - 강제 종료 후 락 파일 정리
- **상태 관리 무결성 확인**
  - 동시 접근 시 상태 일관성
  - 트랜잭션 격리 수준
- **리소스 경합 상황**
  - 디스크 I/O 경합
  - 네트워크 대역폭 경합

### 3. 복구 테스트 (9개)
- **중단된 작업 재개**
  - 클론 중간 중단 후 재개
  - 부분 클론 상태에서 재시작
  - 네트워크 중단 후 재개
- **네트워크 오류 후 복구**
  - 일시적 연결 끊김
  - DNS 해상도 실패
  - 프록시 오류
- **부분 실패 상황 처리**
  - 일부 저장소 클론 실패
  - 권한 변경 중간 발생
  - 디스크 공간 부족 중간 발생
- **데이터 무결성 검증**
  - 중단 후 데이터 손상 여부
  - 체크섬 기반 검증
  - Git 객체 무결성 확인

## 🗂️ 구현 파일 구조

```
test/integration/docker/synclone/
├── docker-compose.yml           # 통합 테스트 환경 구성
├── services/                    # 개별 서비스 설정
│   ├── gitea/
│   │   ├── Dockerfile           # 사용자 정의 Gitea 이미지
│   │   ├── app.ini             # Gitea 설정 파일
│   │   └── init-gitea.sh       # Gitea 초기 설정 스크립트
│   ├── redis/
│   │   └── redis.conf          # Redis 설정
│   └── postgres/
│       ├── init.sql            # 초기 데이터베이스 설정
│       └── seed-data.sql       # 테스트 데이터
├── setup/                       # 환경 설정 스크립트
│   ├── environment-setup.sh     # 전체 환경 설정
│   ├── gitea-setup.sh          # Gitea 서버 초기 설정
│   ├── data-seeder.sh          # 테스트 데이터 생성
│   └── network-setup.sh        # 네트워크 설정
├── tests/                       # 통합 테스트 스크립트
│   ├── api-integration-test.sh  # API 연동 테스트
│   ├── concurrency-test.sh     # 동시성 테스트
│   ├── recovery-test.sh         # 복구 테스트
│   └── end-to-end-test.sh      # 종합 E2E 테스트
├── scenarios/                   # 테스트 시나리오
│   ├── normal-operation.yaml    # 정상 운영 시나리오
│   ├── failure-scenarios.yaml  # 실패 시나리오
│   └── stress-scenarios.yaml   # 스트레스 테스트 시나리오
├── monitoring/                  # 모니터링 도구
│   ├── health-check.sh         # 서비스 상태 확인
│   ├── log-collector.sh        # 로그 수집
│   └── metrics-exporter.sh     # 메트릭 내보내기
├── utils/                       # 유틸리티 도구
│   ├── docker-helpers.sh       # Docker 조작 헬퍼
│   ├── test-data-generator.sh  # 테스트 데이터 생성기
│   └── cleanup-helpers.sh      # 정리 도구
├── integration-test-runner.sh   # 통합 테스트 실행기
└── README.md                   # Phase 3 사용 가이드
```

## 📋 구체적인 구현 계획

### Day 1-2: Docker 환경 구축
- [ ] `docker-compose.yml` 작성 (Gitea, PostgreSQL, Redis)
- [ ] Gitea 서버 커스텀 설정
- [ ] 네트워크 설정 및 포트 매핑
- [ ] 기본 헬스 체크 구현

### Day 3-4: 테스트 데이터 및 환경
- [ ] `data-seeder.sh` - 다양한 저장소/사용자/권한 생성
- [ ] API 키 자동 생성 및 설정
- [ ] 테스트 시나리오 정의
- [ ] 모니터링 도구 구현

### Day 5-6: 핵심 테스트 구현
- [ ] API 연동 테스트 (8개 케이스)
- [ ] 동시성 테스트 (8개 케이스)
- [ ] 복구 테스트 (9개 케이스)
- [ ] 각 테스트의 검증 로직 구현

### Day 7: 통합 및 최적화
- [ ] `integration-test-runner.sh` 통합 실행기
- [ ] CI/CD 파이프라인 통합
- [ ] 성능 최적화 및 안정화
- [ ] 문서화 및 사용 가이드 작성

## ✅ 성공 기준

- [ ] **Docker 환경**에서 완전 자동화된 테스트
- [ ] **실제 Git 서버**와의 연동 검증 (Gitea API 호출)
- [ ] **25개 고급 테스트** 케이스 모두 통과
- [ ] **네트워크/서버 오류** 상황 시뮬레이션 성공
- [ ] **동시성 테스트**에서 데이터 무결성 보장
- [ ] **복구 테스트**에서 100% 데이터 복구 성공
- [ ] **CI/CD 파이프라인** 통합 가능
- [ ] 전체 테스트 실행 시간 **< 45분**

## 🧪 사용 예시

```bash
# 전체 Docker 환경 설정 및 테스트 실행
./test/integration/docker/synclone/integration-test-runner.sh

# Docker 환경만 설정
./test/integration/docker/synclone/setup/environment-setup.sh

# 특정 테스트 카테고리만 실행
./test/integration/docker/synclone/tests/concurrency-test.sh
./test/integration/docker/synclone/tests/recovery-test.sh

# 환경 정리
docker-compose -f test/integration/docker/synclone/docker-compose.yml down -v
```

## 🔧 기술 요구사항

### Docker 환경
- **Docker**: 20.0+
- **Docker Compose**: 2.0+
- **컨테이너 메모리**: 최소 4GB
- **디스크 공간**: 최소 10GB

### 서비스 버전
```yaml
services:
  gitea: gitea/gitea:1.21
  postgres: postgres:15
  redis: redis:7-alpine
  test-runner: golang:1.21
```

### 네트워크 설정
```yaml
networks:
  synclone-test:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## 📊 예상 테스트 결과

### 성능 벤치마크
```
API 호출 응답 시간: < 100ms
동시성 테스트 (10개 프로세스): < 5분
복구 테스트: < 2분
전체 통합 테스트: < 45분
```

### 안정성 지표
```
API 연동 성공률: > 98%
동시성 테스트 데이터 무결성: 100%
복구 테스트 성공률: 100%
메모리 누수: 0건
```

## 🚨 리스크 및 대응

### 주요 리스크
1. **Docker 환경 불안정성**
   - 대응: 헬스 체크 및 재시작 로직
2. **네트워크 타이밍 이슈**
   - 대응: 재시도 로직 및 대기 시간 조정
3. **리소스 부족**
   - 대응: 리소스 모니터링 및 제한 설정

### 성능 최적화
- 컨테이너 리소스 할당 최적화
- 병렬 테스트 실행 개수 조정
- 불필요한 로그 출력 최소화

## 🚀 다음 단계 연계

Phase 3 완료 후:
- **Phase 4**: 지속적 검증 프레임워크
- **운영 환경 적용**: 실제 운영 환경에서의 검증
- **성능 기준선 업데이트**: Docker 환경 기반 성능 기준 설정

---

**작성일**: 2025-08-28  
**예상 시작**: Phase 2 완료 후 (2025-09-07)  
**예상 완료**: 2025-09-14  
**담당자**: DevOps Team, Development Team  
**문서 버전**: 1.0.0