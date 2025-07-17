# PKG

gz pkg ~

gz always-latest를 없애고

- gz
  - pkg
    - all
      - --latest
      - list
    - rbenv
      - --latest
      - list
      - install 3.3.6
    - asdf
      - --latest
      - list
      - plugin
        - list
        - install ~~~
    - brew
      - --latest
      - list
    - apt
      - --latest
      - list
    - port
      - --latest
      - list
    - sdkman
      - --latest
      - list


설치를 지원할지는 고민좀해봐야하고.. 너무 복잡해져서
fix할 버전 이외에는 다 latest로 계속 업데이트 하도록 하는 시나리오
strategy = latest, lts, minor

latest: 무지성 최신버전
lts: 현재 lts버전업
minor: 마이너버전까지만업글

원래는 latest만 하려고 했는데 이렇게 셋은 있어야 할 것 같고
그래서 이름도 always-latest가 아니라 gz pkg로 변경
