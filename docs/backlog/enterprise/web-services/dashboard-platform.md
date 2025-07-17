# 웹 대시보드 플랫폼 기능

## 개요
실시간 모니터링 및 관리를 위한 웹 기반 대시보드 플랫폼

## 제거된 기능

### 1. 웹 서버 및 API
- **명령어**: `gz serve`, `gz serve --port 8080`
- **기능**: REST API 서버 및 웹 대시보드 제공
- **특징**:
  - Gin 웹 프레임워크 기반
  - JWT 인증 시스템
  - RESTful API 엔드포인트
  - CORS 설정 및 보안 헤더

### 2. 실시간 웹소켓 통신
- **기능**: 실시간 데이터 스트리밍 및 양방향 통신
- **특징**:
  - WebSocket 연결 관리
  - 실시간 로그 스트리밍
  - 상태 변경 알림
  - 멀티 클라이언트 지원

### 3. React 기반 프론트엔드
- **기능**: 모던 웹 대시보드 인터페이스
- **특징**:
  - Material-UI 컴포넌트
  - 반응형 디자인
  - 실시간 차트 및 그래프
  - 사용자 인터페이스 커스터마이제이션

### 4. 인증 및 권한 관리
- **기능**: 사용자 인증 및 역할 기반 접근 제어
- **특징**:
  - JWT 토큰 기반 인증
  - 역할 기반 권한 관리
  - 세션 관리
  - 보안 미들웨어

## 제거된 파일 구조

```
web/
├── frontend/                 # React 애플리케이션
│   ├── public/
│   │   ├── index.html
│   │   └── manifest.json
│   ├── src/
│   │   ├── components/       # React 컴포넌트
│   │   │   ├── Dashboard/
│   │   │   ├── Repositories/
│   │   │   ├── Monitoring/
│   │   │   └── Settings/
│   │   ├── hooks/           # React 훅
│   │   ├── services/        # API 서비스
│   │   ├── utils/           # 유틸리티 함수
│   │   ├── App.js
│   │   └── index.js
│   ├── package.json
│   └── webpack.config.js
│
├── backend/                  # Go 웹 서버
│   ├── handlers/            # HTTP 핸들러
│   │   ├── auth.go
│   │   ├── repositories.go
│   │   ├── monitoring.go
│   │   └── websocket.go
│   ├── middleware/          # 미들웨어
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logging.go
│   ├── models/              # 데이터 모델
│   ├── services/            # 비즈니스 로직
│   └── server.go
│
└── assets/                   # 정적 자산
    ├── css/
    ├── js/
    └── images/

cmd/monitoring/
├── monitoring.go             # CLI → 웹 서버로 변경됨
├── websocket.go             # 제거됨
├── auth.go                  # 제거됨
└── instance_manager.go      # 제거됨
```

## 제거된 API 엔드포인트

### 1. 인증 API
```
POST   /api/auth/login       # 사용자 로그인
POST   /api/auth/logout      # 사용자 로그아웃
POST   /api/auth/refresh     # 토큰 갱신
GET    /api/auth/profile     # 사용자 프로필
```

### 2. 저장소 관리 API
```
GET    /api/repositories     # 저장소 목록
GET    /api/repositories/:id # 저장소 상세 정보
POST   /api/repositories     # 저장소 추가
PUT    /api/repositories/:id # 저장소 업데이트
DELETE /api/repositories/:id # 저장소 삭제
POST   /api/repositories/sync # 저장소 동기화
```

### 3. 모니터링 API
```
GET    /api/monitoring/status    # 시스템 상태
GET    /api/monitoring/metrics   # 성능 메트릭
GET    /api/monitoring/logs      # 로그 조회
GET    /api/monitoring/alerts    # 알림 목록
POST   /api/monitoring/alerts    # 알림 생성
```

### 4. 설정 API
```
GET    /api/config              # 설정 조회
PUT    /api/config              # 설정 업데이트
GET    /api/config/schema       # 설정 스키마
POST   /api/config/validate     # 설정 검증
```

### 5. WebSocket 엔드포인트
```
/ws/logs                       # 실시간 로그 스트림
/ws/metrics                    # 실시간 메트릭
/ws/status                     # 실시간 상태 업데이트
/ws/notifications              # 실시간 알림
```

## 제거된 웹 대시보드 기능

### 1. 메인 대시보드
```jsx
// Dashboard.jsx
function Dashboard() {
  return (
    <Grid container spacing={3}>
      <Grid item xs={12} md={6} lg={3}>
        <StatusCard title="Active Repositories" value={repositories.length} />
      </Grid>
      <Grid item xs={12} md={6} lg={3}>
        <StatusCard title="Recent Syncs" value={recentSyncs} />
      </Grid>
      <Grid item xs={12} md={6} lg={3}>
        <StatusCard title="Failed Jobs" value={failedJobs} />
      </Grid>
      <Grid item xs={12} md={6} lg={3}>
        <StatusCard title="System Load" value={systemLoad} />
      </Grid>
      
      <Grid item xs={12}>
        <RecentActivity activities={activities} />
      </Grid>
      
      <Grid item xs={12} md={6}>
        <RepositoryChart data={repositoryData} />
      </Grid>
      
      <Grid item xs={12} md={6}>
        <SystemMetrics metrics={systemMetrics} />
      </Grid>
    </Grid>
  );
}
```

### 2. 저장소 관리 인터페이스
```jsx
// RepositoryManager.jsx
function RepositoryManager() {
  const [repositories, setRepositories] = useState([]);
  const [selectedRepo, setSelectedRepo] = useState(null);
  
  return (
    <Paper>
      <Toolbar>
        <Typography variant="h6">Repositories</Typography>
        <Button onClick={handleAddRepository}>Add Repository</Button>
      </Toolbar>
      
      <DataGrid
        rows={repositories}
        columns={columns}
        onSelectionModelChange={setSelectedRepo}
        pageSize={25}
        checkboxSelection
      />
      
      {selectedRepo && (
        <RepositoryDetails repository={selectedRepo} />
      )}
    </Paper>
  );
}
```

### 3. 실시간 모니터링 컴포넌트
```jsx
// RealTimeMonitoring.jsx
function RealTimeMonitoring() {
  const [metrics, setMetrics] = useState([]);
  const ws = useWebSocket('/ws/metrics');
  
  useEffect(() => {
    ws.onmessage = (event) => {
      const newMetric = JSON.parse(event.data);
      setMetrics(prev => [...prev.slice(-99), newMetric]);
    };
  }, [ws]);
  
  return (
    <Paper>
      <Typography variant="h6">Real-time Metrics</Typography>
      <LineChart data={metrics} />
      <MetricsTable metrics={metrics.slice(-10)} />
    </Paper>
  );
}
```

### 4. 로그 뷰어
```jsx
// LogViewer.jsx
function LogViewer() {
  const [logs, setLogs] = useState([]);
  const [filter, setFilter] = useState('');
  const ws = useWebSocket('/ws/logs');
  
  useEffect(() => {
    ws.onmessage = (event) => {
      const logEntry = JSON.parse(event.data);
      setLogs(prev => [logEntry, ...prev.slice(0, 999)]);
    };
  }, [ws]);
  
  const filteredLogs = logs.filter(log => 
    log.message.includes(filter) || log.level.includes(filter)
  );
  
  return (
    <Paper>
      <TextField
        label="Filter logs"
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        fullWidth
      />
      <VirtualizedList items={filteredLogs} />
    </Paper>
  );
}
```

## 제거된 설정 관리

### 1. 웹 서버 설정
```yaml
web:
  server:
    host: "0.0.0.0"
    port: 8080
    read_timeout: 30s
    write_timeout: 30s
    
  cors:
    allowed_origins: ["http://localhost:3000"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["*"]
    credentials: true
    
  auth:
    jwt_secret: ${JWT_SECRET}
    token_expiry: 24h
    refresh_expiry: 7d
    
  websocket:
    read_buffer_size: 1024
    write_buffer_size: 1024
    check_origin: false
    
  static:
    directory: "./web/dist"
    cache_duration: 24h
```

### 2. 프론트엔드 빌드 설정
```javascript
// webpack.config.js
module.exports = {
  entry: './src/index.js',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: '[name].[contenthash].js',
    publicPath: '/',
  },
  module: {
    rules: [
      {
        test: /\.jsx?$/,
        exclude: /node_modules/,
        use: {
          loader: 'babel-loader',
          options: {
            presets: ['@babel/preset-react']
          }
        }
      }
    ]
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: './public/index.html'
    })
  ],
  devServer: {
    historyApiFallback: true,
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      }
    }
  }
};
```

## 제거된 인증 시스템

### 1. JWT 인증 미들웨어
```go
// middleware/auth.go
func AuthMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        claims, err := validateJWT(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("role", claims.Role)
        c.Next()
    })
}
```

### 2. 사용자 관리
```go
// models/user.go
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Username  string    `json:"username" gorm:"unique"`
    Email     string    `json:"email" gorm:"unique"`
    Password  string    `json:"-"`
    Role      string    `json:"role" gorm:"default:user"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) ValidatePassword(password string) bool {
    return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}
```

## 제거 이유

### 1. 복잡성 증가
- CLI 도구의 핵심 목적과 거리가 먼 기능
- 웹 서버 관리의 추가 복잡성
- 프론트엔드 빌드 및 배포 과정

### 2. 유지보수 부담
- 웹 보안 취약점 관리
- 브라우저 호환성 문제
- UI/UX 지속적 개선 필요

### 3. 사용성 문제
- CLI 사용자들이 웹 인터페이스를 선호하지 않음
- 원격 서버 접근의 보안 우려
- 추가 포트 및 방화벽 설정 필요

## 권장 대안 도구

1. **CLI 기반 모니터링**: htop, ctop, glances
2. **웹 대시보드**: Grafana, Kibana, Prometheus UI
3. **로그 관리**: ELK Stack, Fluentd, Loki
4. **시스템 모니터링**: Nagios, Zabbix, DataDog
5. **프로젝트 관리**: GitHub/GitLab 웹 인터페이스
6. **터미널 UI**: Rich, Bubble Tea, Termui
7. **원격 관리**: SSH + tmux/screen

## 복원 시 고려사항

- 웹 보안 모범 사례 적용
- 반응형 디자인 및 접근성
- API 설계 및 문서화
- 인증/권한 시스템 설계
- 실시간 데이터 처리 최적화
- 프론트엔드 빌드 파이프라인
- 배포 및 서비스 관리
- 사용자 경험 최적화
- 성능 모니터링 및 최적화