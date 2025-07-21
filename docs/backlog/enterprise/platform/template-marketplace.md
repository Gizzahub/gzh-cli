# 템플릿 마켓플레이스 기능

## 개요
프로젝트 템플릿 및 코드 생성기 마켓플레이스

## 제거된 기능

### 1. 템플릿 관리
- **명령어**: `gz template install`, `gz template list`, `gz template remove`
- **기능**: 프로젝트 템플릿 설치 및 관리
- **특징**:
  - 중앙 템플릿 레지스트리
  - 버전 관리
  - 의존성 해결
  - 자동 업데이트

### 2. 프로젝트 생성
- **명령어**: `gz template generate`, `gz template scaffold`
- **기능**: 템플릿 기반 프로젝트 생성
- **특징**:
  - 대화형 프로젝트 설정
  - 조건부 파일 생성
  - 변수 치환
  - 후처리 스크립트

### 3. 템플릿 개발 도구
- **명령어**: `gz template create`, `gz template validate`, `gz template publish`
- **기능**: 템플릿 개발 및 배포 도구
- **특징**:
  - 템플릿 스케폴딩
  - 검증 및 테스트
  - 레지스트리 배포
  - 문서 자동 생성

### 4. 커스터마이제이션
- **명령어**: `gz template customize`, `gz template extend`
- **기능**: 기존 템플릿 커스터마이징 및 확장
- **특징**:
  - 템플릿 상속
  - 부분 오버라이드
  - 믹스인 지원
  - 설정 프로파일

## 사용 예시 (제거 전)

```bash
# 템플릿 검색 및 설치
gz template search --category web --framework react
gz template install react-typescript-app@2.1.0

# 새 프로젝트 생성
gz template generate react-typescript-app \
  --name my-app --author "John Doe" --license MIT

# 커스텀 템플릿 생성
gz template create --name my-company-template \
  --base react-typescript-app

# 템플릿 배포
gz template publish --registry company-internal
```

## 설정 파일 형식

```yaml
templates:
  registries:
    official: https://templates.gzh-manager.io
    company: https://templates.company.com
    local: file:///usr/local/share/gzh-templates

  cache:
    directory: ~/.gzh-templates/cache
    ttl: 24h
    max_size: 1GB

  generation:
    output_directory: ./
    overwrite_policy: prompt
    backup_existing: true

  installed:
    - name: react-typescript-app
      version: 2.1.0
      registry: official
      installed_at: 2024-01-15T10:30:00Z

    - name: go-microservice
      version: 1.5.0
      registry: official
      customizations:
        - add_grpc: true
        - add_swagger: true

    - name: company-backend-api
      version: 1.0.0
      registry: company
      private: true

  defaults:
    author: "Company Dev Team"
    license: "MIT"
    git_init: true
    install_deps: true

  variables:
    global:
      COMPANY_NAME: "ACME Corp"
      DEFAULT_LICENSE: "MIT"

    user:
      AUTHOR_NAME: "John Doe"
      AUTHOR_EMAIL: "john@company.com"
      GITHUB_USERNAME: "johndoe"
```

## 템플릿 구조

### 1. 템플릿 매니페스트
```yaml
name: react-typescript-app
version: 2.1.0
description: React application with TypeScript and modern tooling
author: Template Author <author@example.com>
license: MIT
homepage: https://github.com/example/react-typescript-template

category: web
tags: [react, typescript, vite, testing]
keywords: [frontend, spa, modern]

requirements:
  node_version: ">=16.0.0"
  npm_version: ">=8.0.0"

variables:
  - name: project_name
    type: string
    description: Project name
    required: true
    pattern: "^[a-z][a-z0-9-]*[a-z0-9]$"

  - name: description
    type: string
    description: Project description
    default: "A React TypeScript application"

  - name: author
    type: string
    description: Author name
    default: "{{ .User.Name }}"

  - name: use_router
    type: boolean
    description: Include React Router
    default: true

  - name: ui_framework
    type: choice
    description: UI framework
    choices: [none, material-ui, chakra-ui, antd]
    default: none

  - name: testing_framework
    type: choice
    description: Testing framework
    choices: [jest, vitest]
    default: vitest

conditions:
  - if: "{{ .Variables.use_router }}"
    include: ["src/pages/**", "src/router/**"]

  - if: "{{ eq .Variables.ui_framework 'material-ui' }}"
    include: ["src/theme/**"]
    dependencies: ["@mui/material", "@emotion/react"]

  - if: "{{ eq .Variables.testing_framework 'jest' }}"
    include: ["jest.config.js"]
    exclude: ["vitest.config.ts"]

files:
  - path: "package.json"
    template: true

  - path: "src/App.tsx"
    template: true

  - path: "public/index.html"
    template: true

  - path: ".gitignore"
    static: true

  - path: "README.md"
    template: true

hooks:
  pre_generate:
    - validate_environment

  post_generate:
    - npm_install
    - git_init
    - initial_commit

extends: base-typescript-template
```

### 2. 템플릿 파일 예시
```typescript
// src/App.tsx.tpl
import React from 'react';
{{- if .Variables.use_router }}
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
{{- end }}
{{- if eq .Variables.ui_framework "material-ui" }}
import { ThemeProvider } from '@mui/material/styles';
import theme from './theme';
{{- end }}

function App() {
  return (
    {{- if eq .Variables.ui_framework "material-ui" }}
    <ThemeProvider theme={theme}>
    {{- end }}
      {{- if .Variables.use_router }}
      <Router>
        <Routes>
          <Route path="/" element={<Home />} />
        </Routes>
      </Router>
      {{- else }}
      <div className="App">
        <h1>{{ .Variables.project_name }}</h1>
        <p>{{ .Variables.description }}</p>
      </div>
      {{- end }}
    {{- if eq .Variables.ui_framework "material-ui" }}
    </ThemeProvider>
    {{- end }}
  );
}

export default App;
```

### 3. 패키지 정의
```json
{
  "name": "{{ .Variables.project_name }}",
  "version": "0.1.0",
  "description": "{{ .Variables.description }}",
  "author": "{{ .Variables.author }}",
  "license": "{{ .Variables.license | default "MIT" }}",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    {{- if eq .Variables.testing_framework "jest" }}
    "test": "jest",
    {{- else }}
    "test": "vitest",
    {{- end }}
    "lint": "eslint src --ext .ts,.tsx"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
    {{- if .Variables.use_router }},
    "react-router-dom": "^6.8.0"
    {{- end }}
    {{- if eq .Variables.ui_framework "material-ui" }},
    "@mui/material": "^5.11.0",
    "@emotion/react": "^11.10.5",
    "@emotion/styled": "^11.10.5"
    {{- end }}
  },
  "devDependencies": {
    "@types/react": "^18.0.27",
    "@types/react-dom": "^18.0.10",
    "@vitejs/plugin-react": "^3.1.0",
    "typescript": "^4.9.4",
    "vite": "^4.1.0"
    {{- if eq .Variables.testing_framework "jest" }},
    "jest": "^29.3.1",
    "@types/jest": "^29.2.5"
    {{- else }},
    "vitest": "^0.28.0"
    {{- end }}
  }
}
```

## 템플릿 카테고리

### 1. 웹 애플리케이션
- React/Next.js 앱
- Vue.js/Nuxt.js 앱
- Angular 애플리케이션
- Svelte/SvelteKit 앱
- 정적 사이트 생성기

### 2. 백엔드 서비스
- REST API 서버
- GraphQL 서버
- 마이크로서비스
- 서버리스 함수
- gRPC 서비스

### 3. 모바일 애플리케이션
- React Native 앱
- Flutter 앱
- Ionic 앱
- 네이티브 모바일 앱

### 4. 라이브러리/패키지
- NPM 패키지
- Go 모듈
- Python 패키지
- Rust 크레이트
- Docker 이미지

### 5. 인프라/DevOps
- Kubernetes 매니페스트
- Terraform 모듈
- Ansible 플레이북
- CI/CD 파이프라인
- Docker Compose

### 6. 문서/설정
- README 템플릿
- API 문서
- 프로젝트 설정
- 라이선스 파일
- 기여 가이드

## 고급 기능

### 1. 템플릿 상속
```yaml
# child-template.yaml
extends: parent-template
overrides:
  variables:
    - name: framework_version
      default: "latest"
  files:
    - path: "custom-config.json"
      template: true
```

### 2. 믹스인 지원
```yaml
# main-template.yaml
mixins:
  - testing-mixin
  - linting-mixin
  - docker-mixin

# testing-mixin.yaml
files:
  - path: "tests/"
    recursive: true
hooks:
  post_generate:
    - setup_testing
```

### 3. 조건부 생성
```yaml
conditions:
  - if: "{{ .Variables.include_docker }}"
    files: ["Dockerfile", "docker-compose.yml"]

  - if: "{{ and .Variables.use_database (eq .Variables.db_type 'postgres') }}"
    files: ["migrations/*.sql"]
    dependencies: ["pg"]
```

### 4. 후처리 스크립트
```bash
#!/bin/bash
# hooks/post_generate.sh

echo "Setting up project..."

# 의존성 설치
if [ -f "package.json" ]; then
    npm install
elif [ -f "go.mod" ]; then
    go mod tidy
fi

# Git 초기화
if [ "$GIT_INIT" = "true" ]; then
    git init
    git add .
    git commit -m "Initial commit from template"
fi

echo "Project setup complete!"
```

## 마켓플레이스 기능

### 1. 템플릿 검색
- 카테고리별 분류
- 태그 기반 필터링
- 인기도 및 평점 정렬
- 호환성 검증

### 2. 템플릿 정보
- 상세 설명 및 미리보기
- 사용법 및 예제
- 변경 로그
- 사용자 리뷰

### 3. 품질 관리
- 템플릿 검증
- 보안 스캔
- 베스트 프랙티스 체크
- 커뮤니티 피드백

### 4. 기업용 기능
- 프라이빗 레지스트리
- 조직별 템플릿
- 접근 권한 관리
- 사용 통계

## 통합 예시

### 1. CI/CD 통합
```yaml
# GitHub Actions 워크플로우
name: Template Test
on:
  push:
    paths: ['templates/**']

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Test Template
        run: |
          gz template generate react-app --test-mode
          cd generated-project
          npm test
```

### 2. IDE 통합
```json
{
  "contributes": {
    "commands": [
      {
        "command": "gzh.createFromTemplate",
        "title": "Create from Template"
      }
    ],
    "menus": {
      "explorer/context": [
        {
          "command": "gzh.createFromTemplate",
          "group": "navigation"
        }
      ]
    }
  }
}
```

## 권장 대안 도구

1. **Yeoman**: Node.js 기반 코드 생성기
2. **Cookiecutter**: Python 템플릿 엔진
3. **Plop**: 마이크로 생성기 도구
4. **Hygen**: 코드 생성 도구
5. **GitHub Templates**: GitHub 저장소 템플릿
6. **GitLab Templates**: GitLab 프로젝트 템플릿
7. **Create React App**: React 앱 생성기
8. **Vue CLI**: Vue.js 프로젝트 생성기

## 복원 시 고려사항

- 템플릿 엔진 선택 (Go templates, Handlebars 등)
- 변수 검증 및 타입 시스템
- 파일 시스템 안전성
- 템플릿 저장소 인프라
- 버전 관리 및 호환성
- 보안 스캔 및 검증
- 사용자 경험 최적화
- 커뮤니티 기여 시스템
