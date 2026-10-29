# 배포·출시 문서

출시 명령과 권한·산출물 계약의 정본은 [출시 절차](70-releases.md)입니다.
이 인덱스는 그 명령을 복제하거나 별도의 배포 경로를 정의하지 않습니다.

## 출시 문서

- [출시 절차](70-releases.md): 고정 도구 준비, 비게시 preview, 서명 검증과 기존 bytes 게시
- [출시 준비 체크리스트](71-release-preparation.md): 최종 후보의 통과 조건과 기록
- [출시 노트 초안](72-release-notes.md): 공개 전 검토할 사용자 변경점과 제약
- [변경 이력](../../CHANGELOG.md): 실제 변경점과 출시 상태

## 보안 문서

- [보안 검사](73-security-scanning.md)
- [보안 지침](75-security-guidelines.md)

## 빌드·도구 기준

Go 기준은 [go.mod](../../go.mod)와 저장소 runtime 선언을 따릅니다.
release 도구 pin과 설치 위치는 [.make/tools.mk](../../.make/tools.mk)가 소유합니다.
첫 채널과 실제 게시 순서는 출시 절차를 따르며, 이 문서는 GPG·Docker 등 별도 경로를 추가하지 않습니다.
