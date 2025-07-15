# title: 인프라 및 배포 지원 QA 시나리오

## related_tasks
- /tasks/done/20250712__infrastructure_deployment_support__DONE_20250713.md

## purpose
컨테이너 환경, CI/CD 파이프라인, 클라우드 인프라 관리 기능이 실제 배포 환경에서 정상 동작하는지 검증

## scenarios

### 1. 컨테이너 환경 지원 검증
1. **Dockerfile 자동 생성 테스트**
   ```bash
   # 프로젝트별 Dockerfile 생성
   gz docker generate --project go-web --optimize
   gz docker generate --project node-api --multistage
   
   # 생성된 Dockerfile 검증
   docker build -t test-app .
   docker run --rm test-app
   ```
   - 멀티 스테이지 빌드 템플릿 정상 생성 확인
   - 언어별 최적화 설정 적용 검증
   - 보안 스캔 통합 (trivy, grype) 동작 확인

2. **Helm 차트 생성 및 배포**
   ```bash
   # Helm 차트 생성
   gz k8s helm-generate --app webapp --namespace production
   
   # Kubernetes 클러스터에 배포
   helm install webapp ./webapp-chart
   kubectl get pods -l app=webapp
   ```
   - 차트 템플릿 라이브러리 정상 생성 확인
   - 값 파일 관리 시스템 검증
   - 의존성 차트 자동 처리 동작 확인

3. **K8s 오퍼레이터 기능**
   ```bash
   # CRD 및 오퍼레이터 배포
   gz k8s operator deploy --name gzh-operator
   
   # 커스텀 리소스 생성 및 관리
   kubectl apply -f custom-resource.yaml
   kubectl get gzhmanaged
   ```
   - CRD(Custom Resource Definition) 정상 생성
   - 컨트롤러 구현 (controller-runtime) 동작 검증
   - 리소스 라이프사이클 관리 확인

### 2. CI/CD 파이프라인 통합 검증
1. **GitHub Actions 워크플로우**
   ```bash
   # 워크플로우 생성
   gz ci github-actions --template go-app --security-scan
   
   # 생성된 워크플로우 확인
   cat .github/workflows/ci.yml
   cat .github/workflows/release.yml
   ```
   - 워크플로우 템플릿 라이브러리 정상 생성
   - 재사용 가능한 액션 정의 확인
   - 시크릿 관리 및 보안 스캔 설정 검증

2. **GitLab CI/CD 통합**
   ```bash
   # GitLab CI 설정 생성
   gz ci gitlab --pipeline docker-deploy --registry harbor
   
   # 파이프라인 템플릿 확인
   cat .gitlab-ci.yml
   ```
   - .gitlab-ci.yml 템플릿 정상 생성
   - 파이프라인 단계별 템플릿 확인
   - GitLab Runner 설정 자동화 검증

3. **Jenkins 파이프라인 지원**
   ```bash
   # Jenkinsfile 생성
   gz ci jenkins --type declarative --stages "build,test,deploy"
   
   # 공유 라이브러리 설정
   gz ci jenkins-library --install
   ```
   - Jenkinsfile 생성기 정상 동작
   - 공유 라이브러리 개발 및 배포 확인
   - 플러그인 자동 설치 및 설정 검증

### 3. 클라우드 인프라 관리 검증
1. **Terraform 통합 테스트**
   ```bash
   # Terraform 모듈 생성
   gz terraform generate --provider aws --service ec2,rds,vpc
   
   # 인프라 배포 테스트
   terraform init
   terraform plan
   terraform apply -auto-approve
   ```
   - 클라우드별 모듈 라이브러리 정상 생성
   - Terraform 상태 관리 자동화 확인
   - 계획(plan) 및 적용(apply) 자동화 검증

2. **CloudFormation 템플릿**
   ```bash
   # CloudFormation 스택 생성
   gz aws cloudformation --template webapp-stack
   
   # AWS 리소스 배포
   aws cloudformation create-stack --stack-name test-stack \
     --template-body file://template.yaml
   ```
   - 스택 생성 및 관리 자동화 확인
   - 파라미터 및 출력 처리 검증
   - 변경 세트 미리보기 기능 테스트

3. **Ansible 플레이북 실행**
   ```bash
   # Ansible 플레이북 생성 및 실행
   gz ansible generate --role webserver,database
   ansible-playbook -i inventory site.yml
   ```
   - 역할(Role) 정의 및 관리 확인
   - 인벤토리 파일 자동 생성 검증
   - Ansible Vault 암호화 통합 테스트

### 4. 템플릿 마켓플레이스 검증
1. **커뮤니티 템플릿 공유**
   ```bash
   # 템플릿 업로드
   gz marketplace upload --template ./my-template \
     --category kubernetes --license MIT
   
   # 템플릿 검색 및 다운로드
   gz marketplace search --keyword "go webapp"
   gz marketplace download --id template-123 --output ./downloaded
   ```
   - RESTful API 서버 정상 동작 확인
   - 템플릿 검증 및 승인 프로세스 검증
   - 라이선스 관리 시스템 동작 확인

2. **기업용 프라이빗 마켓플레이스**
   ```bash
   # 엔터프라이즈 인증 및 접근
   gz marketplace login --enterprise --sso
   gz marketplace list --private --org my-company
   ```
   - 역할 기반 접근 제어 (RBAC) 시스템 검증
   - 승인 워크플로우 및 정책 엔진 확인
   - SSO/LDAP 통합 및 사용자 관리 테스트

### 5. 컨테이너 이미지 관리 검증
1. **이미지 자동 빌드 테스트**
   ```bash
   # 멀티 아키텍처 빌드
   gz docker build --platform linux/amd64,linux/arm64 \
     --push --registry docker.io/myorg/app
   
   # 취약점 스캔
   gz docker scan --image myorg/app:latest --output report.json
   ```
   - 멀티 아키텍처 빌드 (amd64, arm64, arm/v7) 확인
   - 이미지 레지스트리 자동 푸시 검증
   - 취약점 스캔 및 SBOM 생성 확인

2. **이미지 최적화 검증**
   ```bash
   # 이미지 최적화 분석
   gz docker optimize --image myorg/app:latest --analyze
   
   # 최적화 적용
   gz docker optimize --image myorg/app:latest --apply \
     --output optimized.dockerfile
   ```
   - 레이어 최적화 및 압축 시스템 확인
   - 베이스 이미지 분석 및 추천 엔진 검증
   - docker-slim 통합 및 외부 도구 지원 확인

### 6. 통합 배포 시나리오
1. **전체 배포 파이프라인**
   ```bash
   # 1. 소스 코드에서 시작
   git clone https://github.com/example/webapp.git
   cd webapp
   
   # 2. CI/CD 파이프라인 설정
   gz ci setup --provider github --deploy kubernetes
   
   # 3. 인프라 준비
   gz terraform apply --env production
   
   # 4. 애플리케이션 배포
   gz deploy --target k8s --namespace production
   ```

2. **멀티 클라우드 배포**
   ```bash
   # AWS 환경 배포
   gz deploy --cloud aws --region us-west-2
   
   # Azure 환경 배포  
   gz deploy --cloud azure --region westus2
   
   # GCP 환경 배포
   gz deploy --cloud gcp --region us-central1
   ```

## expected_results
- **컨테이너**: Dockerfile, Helm 차트, K8s 오퍼레이터 정상 생성 및 배포
- **CI/CD**: GitHub Actions, GitLab CI, Jenkins 파이프라인 정상 동작
- **인프라**: Terraform, CloudFormation, Ansible 자동화 정상 동작
- **마켓플레이스**: 커뮤니티 및 기업용 템플릿 공유 시스템 정상 동작
- **이미지 관리**: 멀티 아키텍처 빌드, 취약점 스캔, 최적화 기능 정상 동작
- **보안**: 모든 단계에서 보안 스캔 및 컴플라이언스 검사 통과
- **멀티 클라우드**: AWS, Azure, GCP 환경에서 일관된 배포 경험

## test_environment
- **클라우드**: AWS, Azure, GCP 테스트 계정
- **컨테이너**: Docker Desktop, Kubernetes 클러스터 (minikube, kind)
- **CI/CD**: GitHub Actions, GitLab CI, Jenkins 서버
- **레지스트리**: Docker Hub, Harbor, ECR, ACR, GCR
- **보안 도구**: Trivy, Grype, Snyk, 코드 서명 인증서
- **모니터링**: Prometheus, Grafana, ELK Stack

## automation_level
- **자동화 가능**: 템플릿 생성, 스크립트 실행, API 호출, 배포 상태 확인
- **수동 검증 필요**: 실제 클라우드 배포, 보안 정책 검증, 복합 워크플로우

## tags
[qa], [deployment], [infrastructure], [ci-cd], [containers], [cloud], [automated], [manual]