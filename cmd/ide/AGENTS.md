# AGENTS.md - ide (IDE 관리)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 🎯 모듈 특성

**ide**는 다양한 IDE(JetBrains, VS Code 계열 등)를 감지하고 관리하는 모듈입니다.

### 핵심 기능
- IDE 자동 감지 (JetBrains, VS Code, Cursor 등)
- IDE 실행 및 프로젝트 열기
- JetBrains 설정 모니터링
- 설정 동기화 문제 해결 (fix-sync)

## ⚠️ 개발 시 주의사항

### 1. 크로스 플랫폼 IDE 경로
```go
// ✅ 플랫폼별 IDE 경로 처리
func (d *IDEDetector) getIDEPaths() map[string][]string {
    switch runtime.GOOS {
    case "darwin":
        return map[string][]string{
            "vscode": {"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"},
            "pycharm": {"/Applications/PyCharm.app/Contents/bin/pycharm"},
        }
    case "linux":
        return map[string][]string{
            "vscode": {"/usr/bin/code", "/snap/bin/code"},
            "pycharm": {"/opt/pycharm/bin/pycharm.sh"},
        }
    case "windows":
        return map[string][]string{
            "vscode": {"C:\\Users\\%USERNAME%\\AppData\\Local\\Programs\\Microsoft VS Code\\bin\\code.cmd"},
        }
    }
}
```

### 2. JetBrains 설정 디렉토리 처리
```go
// ✅ JetBrains 설정 경로 관리
func (j *JetBrainsManager) getConfigPaths() ([]string, error) {
    homeDir, _ := os.UserHomeDir()

    switch runtime.GOOS {
    case "darwin":
        return []string{
            filepath.Join(homeDir, "Library/Application Support/JetBrains"),
            filepath.Join(homeDir, "Library/Preferences"),
        }, nil
    case "linux":
        return []string{
            filepath.Join(homeDir, ".config/JetBrains"),
            filepath.Join(homeDir, ".local/share/JetBrains"),
        }, nil
    case "windows":
        return []string{
            filepath.Join(os.Getenv("APPDATA"), "JetBrains"),
        }, nil
    }
}
```

### 3. 안전한 IDE 실행
```go
// ✅ IDE 실행 안전성
func (i *IDELauncher) LaunchIDE(ideName, projectPath string) error {
    // 프로젝트 경로 검증
    if !i.isValidProjectPath(projectPath) {
        return fmt.Errorf("invalid project path: %s", projectPath)
    }

    // IDE 실행 파일 존재 확인
    exe, err := i.findIDEExecutable(ideName)
    if err != nil {
        return fmt.Errorf("IDE not found: %s", ideName)
    }

    // 백그라운드 실행
    cmd := exec.Command(exe, projectPath)
    return cmd.Start() // Run()이 아닌 Start() 사용
}
```

## 🧪 테스트 요구사항

- **IDE 버전별 테스트**: 다양한 IDE 버전 호환성
- **설정 파일 처리**: 손상된 설정 파일 복구
- **동시 실행**: 여러 IDE 동시 실행 시나리오
- **경로 문제**: 공백이 포함된 경로 처리

**핵심**: IDE는 사용자의 개발 환경이므로 설정 손상을 방지하고 안전한 실행을 보장해야 합니다.
