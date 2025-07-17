# Terraform 관리 기능

## 개요
Infrastructure as Code(IaC)를 위한 Terraform 통합 관리 기능

## 제거된 기능

### 1. Terraform 계획 및 적용
- **명령어**: `gz terraform plan`, `gz terraform apply`
- **기능**: Terraform 계획 생성 및 인프라 변경 적용
- **특징**:
  - 다중 환경 지원 (dev, staging, prod)
  - 변경사항 미리보기 및 승인 워크플로우
  - 상태 파일 백업 및 복원
  - 롤백 기능

### 2. Terraform 코드 생성
- **명령어**: `gz terraform generate`
- **기능**: 기존 인프라로부터 Terraform 코드 자동 생성
- **특징**:
  - AWS, GCP, Azure 리소스 스캔
  - HCL 코드 자동 생성
  - 모듈화된 구조 생성
  - 베스트 프랙티스 적용

### 3. 상태 관리
- **명령어**: `gz terraform state`
- **기능**: Terraform 상태 파일 관리 및 조작
- **특징**:
  - 원격 상태 백엔드 설정
  - 상태 파일 마이그레이션
  - 리소스 임포트/제거
  - 상태 잠금 관리

## 사용 예시 (제거 전)

```bash
# 개발 환경 계획 생성
gz terraform plan --env dev --workspace myapp

# 프로덕션 환경 적용
gz terraform apply --env prod --auto-approve

# 기존 인프라에서 코드 생성
gz terraform generate --provider aws --region us-west-2

# 상태 파일 백업
gz terraform state backup --remote s3://terraform-state-bucket
```

## 설정 파일 형식

```yaml
terraform:
  workspaces:
    - name: myapp-dev
      environment: development
      backend:
        type: s3
        bucket: terraform-state-dev
        key: myapp/terraform.tfstate
        region: us-west-2
    - name: myapp-prod
      environment: production
      backend:
        type: s3
        bucket: terraform-state-prod
        key: myapp/terraform.tfstate
        region: us-west-2
  
  providers:
    aws:
      regions: [us-west-2, us-east-1]
      profiles: [default, prod]
    
  modules:
    auto_generate: true
    structure: modular
    
  approval:
    required_for: [production]
    reviewers: [devops-team]
```

## 권장 대안 도구

1. **Terraform CLI**: 공식 Terraform 명령줄 도구
2. **Terragrunt**: Terraform 래퍼 도구
3. **Terraform Cloud**: HashiCorp의 관리형 서비스
4. **Atlantis**: GitOps 기반 Terraform 자동화
5. **Spacelift**: 엔터프라이즈 Terraform 플랫폼

## 복원 시 고려사항

- Terraform 버전 호환성
- 상태 파일 마이그레이션 전략
- 보안 및 권한 관리
- CI/CD 파이프라인 통합