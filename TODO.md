# TODO to bulk-clone

## 중요!!
테스트 할 때 Gizzahub에다 하지 않기 ^^
동작 잘 하는거 확인했다 ^^b al

## Bulk Clone
- [x] bulk-clone cli
  - [x] -o organization name: `-o ScriptonBasestar`
  - [x] -t targetPath: `-t $HOME/mywork/ScriptonBasestar`
  - [x] -p proto: `-p https` or `-p ssh`
  - [x] -a auth: `-a token` 토큰이 있으면 private repo도 clone
  - [x] -s strategy: `-s reset` or `-s pull` or `-s fetch`
- [x] bulk-clone.yaml 설정파일 지원
- [x] bulk-clone.yaml override bulk-clone.home.yaml bulk-clone.work.yaml. kustomize 이용
- [x] bulk-clone.yaml 설정파일 예시
- [x] bulk-clone.yaml schema

### Github Org
- [x] gituhb org bulk-clone cli

### Gitlab Org
- [x] gitlab group bulk-clone cli

### Ssh config
- [x] ssh config 설정
- [x] config 설정 만들어내기

## Always Latest
- [x] asdf
- [x] brew
- [x] sdkman
- [x] port
- [x] apt-get
- [x] rbenv

업데이트 방식
- minor latest
- major latest

## IDE
Jetbrains 설정변경 감지 https://github.com/fsnotify/fsnotify
리눅스 (Linux)
경로: ~/.config/JetBrains/<Product><Version>/
예: ~/.config/JetBrains/IntelliJIdea2023.2/
맥OS (MacOS)
경로: ~/Library/Application Support/JetBrains/<Product><Version>/
예: ~/Library/Application Support/JetBrains/IntelliJIdea2023.2/
윈도우 (Windows)
경로: %APPDATA%\JetBrains\<Product><Version>\
예: C:\Users\<YourUserName>\AppData\Roaming\JetBrains\IntelliJIdea2023.2\

setting sync 오류나서 이런것들 강제 수정
- [moved to SKIP.md] ~/.config/JetBrains/PyCharm2024.3/settingsSync/options/filetypes.xml

## 개발환경 Save & Load
- [x] kubeconfig
- [ ] docker config
- [ ] aws config
- [ ] aws credentials
- [ ] gcloud config
- [ ] gcloud credentials
- [ ] ssh config
- ...

## 네트워크 옮길 때
- [ ] daemon 모니터링
- [ ] hook wifi change event? -> action
- [ ] action: vpn, dns, proxy, host 등 변경

## Github org, repo 기본설정
terraform으로 하는게 나을지도 모름
github action으로 할까 했는데 좀 안맞는 것 같아서 여기로 이동

참고프로젝트

- https://github.com/actions/hello-world-docker-action
- https://github.com/actions/typescript-action
- https://github.com/actions/hello-world-javascript-action
- https://github.com/actions/javascript-action
- https://github.com/actions/starter-workflows

참고독s
- https://docs.github.com/ko/actions/security-for-github-actions/security-guides/automatic-token-authentication
- https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#update-a-repository

env
https://stackoverflow.com/questions/73955908/how-to-use-env-variable-as-default-value-for-input-in-github-actions

```
        const repoUpdateResult = octokit.repos.update({
          owner: repoOwner,
          repo: repo.name,

          name: repo.name,
          // description: repo.description,
          // homepage: repo.homepage,
          private: repoMeta.private,
          visibility: repoMeta.visibility,
          security_and_analysis: repoMeta.security_and_analysis,

          has_issues: repoMeta.has_issues,
          has_projects: repoMeta.has_projects,
          has_wiki: repoMeta.has_wiki,

          default_branch: repo.default_branch,

          allow_squash_merge: repoMeta.allow_squash_merge,
          allow_merge_commit: repoMeta.allow_merge_commit,
          allow_rebase_merge: repoMeta.allow_rebase_merge,

          delete_branch_on_merge: repoMeta.delete_branch_on_merge,

          allow_update_branch: repoMeta.allow_update_branch,

          use_squash_pr_title_as_default: repoMeta.use_squash_pr_title_as_default,

          squash_merge_commit_title: repoMeta.squash_merge_commit_title,
          squash_merge_commit_message: repoMeta.squash_merge_commit_message,

          merge_commit_title: repoMeta.merge_commit_title,
          merge_commit_message: repoMeta.merge_commit_message,

          archived: repoMeta.archived,
          allow_forking: repoMeta.allow_forking,
          allow_auto_merge: repoMeta.allow_auto_merge,

          web_commit_signoff_required: repoMeta.web_commit_signoff_required,
        })
```
### 토큰 권한 문서
https://docs.github.com/ko/rest/authentication/permissions-required-for-github-apps?apiVersion=2022-11-28
https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication
