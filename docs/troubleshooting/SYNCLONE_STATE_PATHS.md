# Synclone 상태 파일: Windows의 긴 조직 이름

## 증상과 원인

긴 조직 이름으로 상태를 저장할 때 Windows에서 `failed to inspect state component`
오류가 발생할 수 있었다. `TestStatePathSafetyLongAndUnicodeSubgroupRoundTrip`의
300자 조직 이름이 master의 Windows CI에서 이 문제를 재현했다.

상태는 고정 길이 해시 파일명으로 저장하지만, 이전 버전의 상태와 중복 여부를
확인하려고 옛 파일명도 조회한다. Windows의 `ERROR_INVALID_NAME`(123)은 불가
문자뿐 아니라 너무 긴 단일 파일명 구성 요소에도 반환될 수 있다. 이 오류를
길이와 무관하게 무시하면 확인하지 못한 이전 상태를 덮어쓰거나 숨길 수 있다.

## 처리 기준

열린 부모 디렉터리의 핸들로 실제 파일시스템의 최대 구성 요소 길이를 조회하고,
UTF-16 코드 단위 길이가 이 제한을 초과하는 optional legacy 후보만 없는 것으로
취급한다. canonical 경로의 오류, 길이와 무관한 잘못된 이름, 권한 오류 및 길이
조회 실패는 계속 오류로 반환한다. JSON 형식과 해시 파일명, 루트 제한, symlink·
손상·중복 상태의 보존 규칙은 유지한다. 기존 장문 legacy 파일을 읽는 동작도 유지한다.

수동 Actions 실행은 작업 브랜치에서도 Ubuntu·macOS·Windows를 모두 검사한다.
일반 작업 브랜치 push와 PR은 기존 Ubuntu 범위를 유지한다.

## 지원 범위와 제한

Windows 검증은 GitHub Actions의 로컬 파일시스템에서 수행한다. SMB 공유에 대한
동작은 검증하지 않는다. Microsoft는 `GetVolumeInformationByHandleW`의 SMB 지원을
제한하므로, 해당 조회가 실패하는 환경에서는 원래 조회 오류와 길이 확인 오류를
함께 반환한다. 실패를 숨기거나 고정된 길이를 추정하지 않는다.

공식 계약: [GetVolumeInformationByHandleW](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getvolumeinformationbyhandlew).

## 회귀 검증

기존 긴 이름·Unicode 왕복과 legacy 호환성·중복·symlink 보존 테스트를 유지한다.
Windows 전용 테스트는 실제 긴 이름 오류, UTF-16 길이 경계, canonical 조회 오류와
길이 이외의 잘못된 이름이 숨겨지지 않는지 확인한다. 전체 검증은 `make check`와
작업 브랜치의 수동 Main CI/CD Pipeline 실행을 사용한다.
