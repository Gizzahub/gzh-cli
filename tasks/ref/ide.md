제가 이해한 내용은 **Ubuntu에서 JetBrains IDE(Pycharm, IntelliJ 등)를 CLI 명령어로 실행할 수 있도록 설정하는 방법**을 알고 싶으신 거군요. 예를 들어 `pycharm .` 이라고 입력했을 때 현재 디렉토리를 PyCharm으로 열 수 있도록요.

---

## JetBrains CLI 실행 설정 방법 (Ubuntu)

### 1. IDE에서 "Create Command-Line Launcher…" 사용

1. JetBrains IDE(Pycharm, IntelliJ, WebStorm 등) 실행
2. 상단 메뉴에서 **Tools → Create Command-Line Launcher…** 클릭
3. 원하는 명령어 이름 입력 (예: `pycharm`)

   * 기본적으로 `/usr/local/bin/pycharm` 같은 위치에 링크가 만들어집니다.
4. 터미널에서 바로 실행 가능:

   ```bash
   pycharm .
   ```

👉 이렇게 하면 `pycharm .`, `idea .`, `webstorm .` 처럼 원하는 IDE 명령어로 현재 디렉토리를 열 수 있습니다.

---

### 2. 수동으로 심볼릭 링크 생성

만약 GUI 메뉴가 없다면, 직접 심볼릭 링크를 만들 수도 있습니다.

예: PyCharm Community 설치된 경우:

```bash
sudo ln -s ~/Downloads/pycharm-community-2025.1/bin/pycharm.sh /usr/local/bin/pycharm
```

그 후 실행:

```bash
pycharm .
```

---

### 3. JetBrains Toolbox를 쓰는 경우

Toolbox App에서 IDE를 설치했다면 실행 파일이 보통 다음 경로에 있습니다:

```bash
~/.local/share/JetBrains/Toolbox/apps/PyCharm-C/ch-0/241.14494.240/bin/pycharm.sh
```

링크를 걸어주면 됩니다:

```bash
sudo ln -s ~/.local/share/JetBrains/Toolbox/apps/PyCharm-C/ch-0/*/bin/pycharm.sh /usr/local/bin/pycharm
```

---

✅ 정리하면:

* 가장 편한 방법은 IDE 메뉴의 **"Create Command-Line Launcher…"** 기능을 쓰는 것.
* 수동 방법은 `pycharm.sh` 실행 파일을 `/usr/local/bin` 같은 곳에 링크 걸어두는 것.

---

원하시는 게 **모든 JetBrains IDE (예: idea, webstorm, goland 등)도 CLI로 실행**되도록 하는 설정 방법도 같이 알려드릴까요?
