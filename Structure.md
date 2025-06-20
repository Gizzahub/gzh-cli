.
├── cmd/                # CLI entrypoint 및 main.go
├── internal/
│   ├── config/         # 설정 파일 파싱 및 관리
│   ├── git/            # git 연동 로직
│   ├── sync/           # 동기화 로직
│   ├── auth/           # 인증 모듈
│   ├── logger/         # 로그 및 출력 포매팅
│   └── utils/          # 기타 보조 함수
├── pkg/                # 외부에서 사용할 수 있는 패키지
├── scripts/            # 빌드 및 배포 스크립트
├── tools/              # 의존성 도구
├── docs/
│   ├── FunctionalDetails.md
│   ├── Usage.md
│   └── Structure.md
├── go.mod
├── go.sum
└── README.md
