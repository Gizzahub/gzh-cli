"""native 출시 프로파일 하나를 수집하고 게시 없이 실행 증거를 보존한다."""

import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import time


def main():
    root = Path.cwd().resolve()
    expected_sha, cohort, tags = sys.argv[1:]
    git = lambda *args: subprocess.check_output(["git", *args], text=True).strip()
    if git("rev-parse", "HEAD") != expected_sha:
        raise RuntimeError("checkout differs from the requested immutable source")
    if git("status", "--porcelain"):
        raise RuntimeError("coverage collection requires a clean source")
    packet = Path(os.environ["RUNNER_TEMP"]) / ("core-" + cohort + "-" + (tags or "default"))
    packet.mkdir()
    home = packet / "home"
    home.mkdir()
    temporary = packet / "tmp"
    temporary.mkdir()
    for name in ("config", "credentials"):
        (home / name).write_bytes(b"")
    removed = [key for key in os.environ if key.startswith("AWS_") or
               any(word in key.upper() for word in ("TOKEN", "SECRET", "PASSWORD", "CREDENTIAL")) or
               key in ("SSH_AUTH_SOCK", "KUBECONFIG", "DOCKER_CONFIG", "GOOGLE_APPLICATION_CREDENTIALS",
                       "AZURE_CONFIG_DIR", "GNUPGHOME", "SOPS_AGE_KEY_FILE")]
    env = {key: value for key, value in os.environ.items() if key not in removed}
    # runner의 사용자 설정 경로를 바꾸기 전에 Go 캐시 경로만 보존한다.
    for key in ("GOPATH", "GOMODCACHE", "GOCACHE"):
        env[key] = subprocess.check_output(["go", "env", key], text=True).strip()
    env.update(HOME=str(home), USERPROFILE=str(home), XDG_CONFIG_HOME=str(home / "xdg"),
               APPDATA=str(home / "appdata"), LOCALAPPDATA=str(home / "localappdata"),
               TMPDIR=str(temporary), TMP=str(temporary), TEMP=str(temporary),
               AWS_CONFIG_FILE=str(home / "config"), AWS_SHARED_CREDENTIALS_FILE=str(home / "credentials"),
               AWS_EC2_METADATA_DISABLED="true", GOENV="off", GOTOOLCHAIN="local", GOWORK="off",
               GOMAXPROCS="2", GOFLAGS="-p=2", CGO_ENABLED="0")
    native = json.loads(subprocess.check_output(["go", "env", "-json", "GOOS", "GOARCH", "GOVERSION"], env=env))
    if native["GOOS"] + "-" + native["GOARCH"] != cohort or native["GOVERSION"] != "go1.27.3":
        raise RuntimeError("runner does not match the pinned native cohort/toolchain")
    artifacts = root / "tmp/core-coverage-artifacts"
    artifacts.mkdir(parents=True)
    logs = root / "tmp/core-coverage-execution"
    logs.mkdir(parents=True)
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    record = {"source_commit": expected_sha, "source_tree": git("rev-parse", "HEAD^{tree}"),
              "native": native, "removed_environment_names": sorted(removed),
              "collector_sha256": sha(Path(__file__)), "python": sys.version,
              "isolation": "owned configuration and empty credentials; no OS/network sandbox claim",
              "commands": []}

    def run(argv, label):
        started = time.time()
        log = logs / (label + ".log")
        result = {"argv": argv, "exit": None, "started_unix": started}
        try:
            with log.open("wb") as stream:
                process = subprocess.run(argv, env=env, stdout=stream, stderr=subprocess.STDOUT, timeout=900)
            result.update(exit=process.returncode, termination="exited")
        except subprocess.TimeoutExpired:
            result["termination"] = "timeout"
            raise
        except OSError as error:
            result.update(termination="launch_error", error=str(error))
            raise
        finally:
            result.update(finished_unix=time.time(), raw_log_sha256=sha(log) if log.exists() else None)
            record["commands"].append(result)
            (logs / "execution.json").write_text(json.dumps(record, indent=2) + "\n")
        if result["exit"]:
            raise RuntimeError(label + " failed; retained raw execution log")

    manifest = "tmp/coveragecontract/coverage-contract.json"
    references = "tmp/coveragecontract/profiles"
    run(["make", "coverage-contract-instrument"], "instrument")
    key = cohort + "__" + ("+".join(tags.split(",")) if tags else "default")
    profile = artifacts / (key + ".out")
    command = ["go", "test", "-count=1", "-parallel=1", "-covermode=set", "-coverprofile=" + str(profile)]
    if tags:
        command.append("-tags=" + tags)
    command.append("./...")
    run(command, "go-test")
    reference = Path(references) / (key + ".expected")
    run(["go", "run", "./scripts/coveragecontract", "bind", "-root", ".", "-manifest", manifest,
         "-cohort", cohort, "-tags", tags, "-profile", str(profile), "-expected", str(reference),
         "-out", str(artifacts / (key + ".provenance.json"))], "bind")
    for path in (reference, Path(str(reference) + ".provenance.json")):
        shutil.copyfile(path, artifacts / path.name)
    record["artifacts_sha256"] = {path.name: sha(path) for path in sorted(artifacts.iterdir())}
    if git("rev-parse", "HEAD") != expected_sha or git("status", "--porcelain"):
        raise RuntimeError("source changed during coverage collection")
    (logs / "execution.json").write_text(json.dumps(record, indent=2) + "\n")


if __name__ == "__main__":
    main()
