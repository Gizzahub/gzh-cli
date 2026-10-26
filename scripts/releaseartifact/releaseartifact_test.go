package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const testSHA = "1234567890123456789012345678901234567890"

type fixture struct {
	o              options
	docs           map[string][]byte
	artifacts      []artifact
	commands       [][]string
	signatureError error
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func archiveBytes(t *testing.T, p platform, docs map[string][]byte, attack string) []byte {
	t.Helper()
	files := map[string][]byte{"gz": []byte("binary"), "manpages/gzh-manager.1.gz": []byte("manual"), "completions/gzh-manager.bash": []byte("bash"), "completions/gzh-manager.zsh": []byte("zsh"), "completions/gzh-manager.fish": []byte("fish")}
	if p.OS == "windows" {
		delete(files, "gz")
		files["gz.exe"] = []byte("binary")
	}
	for n, b := range docs {
		files[n] = b
	}
	if attack == "document" {
		files["LICENSE"] = []byte("wrong source")
	}
	if attack == "traversal" {
		files["../escaped"] = []byte("bad")
	}
	if attack == "missing" {
		delete(files, "completions/gzh-manager.fish")
	}
	var output bytes.Buffer
	if p.OS == "windows" {
		w := zip.NewWriter(&output)
		for n, b := range files {
			f, err := w.Create(n)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.Write(b); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		gz := gzip.NewWriter(&output)
		w := tar.NewWriter(gz)
		for n, b := range files {
			if err := w.WriteHeader(&tar.Header{Name: n, Size: int64(len(b)), Mode: 0o644, Typeflag: tar.TypeReg}); err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write(b); err != nil {
				t.Fatal(err)
			}
		}
		if attack == "link" {
			if err := w.WriteHeader(&tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "../escape"}); err != nil {
				t.Fatal(err)
			}
		}
		if attack == "duplicate" {
			if err := w.WriteHeader(&tar.Header{Name: "LICENSE", Size: int64(len(docs["LICENSE"])), Mode: 0o644, Typeflag: tar.TypeReg}); err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write(docs["LICENSE"]); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return output.Bytes()
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "dist")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	f := &fixture{o: options{Dir: dir, Tag: "v0.2.0", SHA: testSHA, Repo: root, Cosign: filepath.Join(root, "cosign")}, docs: map[string][]byte{"LICENSE": []byte("MIT\n"), "README.md": []byte("readme\n"), "CHANGELOG.md": []byte("changelog\n")}}
	writeFile(t, f.o.Cosign, []byte("fake executable placeholder"))
	writeJSON(t, filepath.Join(dir, "metadata.json"), metadata{Project: "gzh-cli", Tag: f.o.Tag, Version: "0.2.0", Commit: testSHA})
	var cask strings.Builder
	fmt.Fprintln(&cask, "cask \"gz\" do\n  version \"0.2.0\"\n  binary \"gz\"")
	for _, p := range platforms {
		n := "gzh-cli_0.2.0_" + p.OS + "_" + p.Arch
		if p.ARM != "" {
			n += "v" + p.ARM
		}
		if p.OS == "windows" {
			n += ".zip"
		} else {
			n += ".tar.gz"
		}
		b := archiveBytes(t, p, f.docs, "")
		writeFile(t, filepath.Join(dir, n), b)
		f.artifacts = append(f.artifacts, artifact{Name: n, Path: filepath.ToSlash(filepath.Join("dist", n)), Type: "Archive", OS: p.OS, Arch: p.Arch, ARM: p.ARM})
		sbom := n + ".sbom.json"
		writeFile(t, filepath.Join(dir, sbom), []byte(`{"schema":"sbom"}`))
		f.artifacts = append(f.artifacts, artifact{Name: sbom, Path: sbom, Type: "SBOM"})
		if p.OS != "windows" && p.ARM == "" {
			fmt.Fprintf(&cask, " sha256 \"%s\"\n url \"https://github.com/Gizzahub/gzh-cli/releases/download/v#{version}/%s\"\n", digest(b), n)
		}
	}
	fmt.Fprintln(&cask, "end")
	writeFile(t, filepath.Join(dir, "gz.rb"), []byte(cask.String()))
	f.artifacts = append(f.artifacts, artifact{Name: "gz.rb", Path: "gz.rb", Type: "Homebrew Cask"})
	// 실제 GoReleaser의 dist/... 상대 경로는 fixture 저장소 기준으로 해석한다.
	for i := range f.artifacts {
		f.artifacts[i].Path = f.artifacts[i].Name
	}
	f.saveArtifacts(t)
	writeFile(t, filepath.Join(dir, "CHANGELOG.md"), []byte("generated release body\n"))
	return f
}

func (f *fixture) saveArtifacts(t *testing.T) {
	t.Helper()
	writeJSON(t, filepath.Join(f.o.Dir, "artifacts.json"), f.artifacts)
}

func (f *fixture) run(_ context.Context, executable string, args, env []string, cwd string) ([]byte, error) {
	f.commands = append(f.commands, append([]string{executable}, args...))
	if executable == "git" {
		if args[0] == "show" {
			_, name, _ := strings.Cut(args[1], ":")
			return f.docs[name], nil
		}
		switch args[0] {
		case "rev-parse":
			return []byte(testSHA), nil
		case "status":
			return nil, nil
		case "ls-remote":
			return []byte(strings.Repeat("a", 40) + "\trefs/tags/v0.2.0\n" + testSHA + "\trefs/tags/v0.2.0^{}\n"), nil
		}
	}
	if executable == f.o.Cosign {
		for _, v := range env {
			if strings.HasPrefix(v, "COSIGN_") || strings.HasPrefix(v, "SIGSTORE_") {
				return nil, errors.New("ambient trust override")
			}
		}
		if args[0] == "version" {
			return []byte(`{"gitVersion":"v3.1.3"}`), nil
		}
		return nil, f.signatureError
	}
	return nil, errors.New("unexpected command")
}

func (f *fixture) stage(t *testing.T) {
	t.Helper()
	if err := prepare(context.Background(), f.o, f.run); err != nil {
		t.Fatal(err)
	}
	f.o.Dir = filepath.Join(f.o.Dir, "publish")
	writeFile(t, filepath.Join(f.o.Dir, bundleName), []byte(`{"bundle":"fake"}`))
}

func TestPrepareAndVerifyBoundSignedInventory(t *testing.T) {
	f := newFixture(t)
	f.stage(t)
	t.Setenv("COSIGN_FULCIO_URL", "https://untrusted.invalid")
	t.Setenv("SIGSTORE_ROOT_FILE", "bad")
	files, m, err := verify(context.Background(), f.o, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Assets) != 16 || len(files) != 19 {
		t.Fatalf("unexpected complete inventory assets=%d files=%d", len(m.Assets), len(files))
	}
	last := f.commands[len(f.commands)-1]
	private := filepath.Dir(last[len(last)-1])
	if private == f.o.Dir {
		t.Fatal("cosign read mutable stage")
	}
	if _, err := os.Stat(private); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("private verification files were not cleaned")
	}
	want := []string{f.o.Cosign, "verify-blob", "--bundle", filepath.Join(private, bundleName), "--certificate-identity", "https://github.com/Gizzahub/gzh-cli/.github/workflows/release.yml@refs/tags/v0.2.0", "--certificate-oidc-issuer", "https://token.actions.githubusercontent.com", "--certificate-github-workflow-sha", testSHA, "--certificate-github-workflow-repository", "Gizzahub/gzh-cli", "--certificate-github-workflow-ref", "refs/tags/v0.2.0", "--certificate-github-workflow-trigger", "push", filepath.Join(private, checksumName)}
	if !reflect.DeepEqual(last, want) {
		t.Fatalf("signer constraints differ: %v", last)
	}
}

func TestPrepareRejectsUnsafeOrIncompleteInputs(t *testing.T) {
	for _, kind := range []string{"missing-body", "wrong-sha", "wrong-tag", "missing-sbom", "duplicate", "traversal", "symlink", "missing-platform", "cask"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			switch kind {
			case "missing-body":
				if err := os.Remove(filepath.Join(f.o.Dir, "CHANGELOG.md")); err != nil {
					t.Fatal(err)
				}
			case "wrong-sha":
				f.o.SHA = strings.Repeat("b", 40)
			case "wrong-tag":
				f.o.Tag = "v0.2.1"
			case "missing-sbom":
				f.artifacts = append(f.artifacts[:1], f.artifacts[2:]...)
			case "duplicate":
				f.artifacts = append(f.artifacts, f.artifacts[0])
			case "traversal":
				f.artifacts[0].Path = "../../escape.tar.gz"
			case "symlink":
				name := filepath.Join(f.o.Dir, f.artifacts[0].Name)
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(f.o.Cosign, name); err != nil {
					t.Fatal(err)
				}
			case "missing-platform":
				f.artifacts = f.artifacts[2:]
			case "cask":
				writeFile(t, filepath.Join(f.o.Dir, "gz.rb"), []byte("invalid"))
			}
			f.saveArtifacts(t)
			if err := prepare(context.Background(), f.o, f.run); err == nil {
				t.Fatal("unsafe input accepted")
			}
			if _, err := os.Stat(filepath.Join(f.o.Dir, "publish")); !os.IsNotExist(err) {
				t.Fatal("failed preflight created stage")
			}
		})
	}
}

func TestArchiveIntegrityContract(t *testing.T) {
	for _, attack := range []string{"document", "traversal", "missing", "link", "duplicate"} {
		t.Run(attack, func(t *testing.T) {
			f := newFixture(t)
			p := platform{OS: "linux", Arch: "amd64"}
			if err := validateArchive(asset{Name: "bad.tar.gz", Platform: &p}, archiveBytes(t, p, f.docs, attack), f.docs); err == nil {
				t.Fatal("unsafe archive accepted")
			}
		})
	}
}

func TestZIPMemberSizeCannotOverflowSignedLimit(t *testing.T) {
	for _, size := range []uint64{uint64(maxFile) + 1, math.MaxUint64} {
		t.Run(fmt.Sprintf("size-%d", size), func(t *testing.T) {
			var b bytes.Buffer
			w := zip.NewWriter(&b)
			if _, err := w.CreateRaw(&zip.FileHeader{Name: "gz.exe", Method: zip.Store, UncompressedSize64: size}); err != nil {
				t.Fatal(err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			z, err := zip.NewReader(bytes.NewReader(b.Bytes()), int64(b.Len()))
			if err != nil {
				t.Fatal(err)
			}
			if len(z.File) != 1 || z.File[0].UncompressedSize64 != size {
				t.Fatal("overflow fixture did not preserve declared size")
			}
			v := archiveValidator{seen: make(map[string]bool), binary: "gz.exe"}
			if err := validateZIPMember(z.File[0], &v); err == nil || !strings.Contains(err.Error(), "unbounded archive member") {
				t.Fatalf("oversized ZIP member accepted or wrong failure: %v", err)
			}
		})
	}
}

func TestVerifyRejectsTamperSignatureAndIdentity(t *testing.T) {
	for _, kind := range []string{"asset", "checksums", "bundle", "extra", "manifest", "tag", "sha", "signature", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			f.stage(t)
			switch kind {
			case "asset":
				writeFile(t, filepath.Join(f.o.Dir, "gz.rb"), []byte("tampered"))
			case "checksums":
				writeFile(t, filepath.Join(f.o.Dir, checksumName), []byte("tampered"))
			case "bundle":
				if err := os.Remove(filepath.Join(f.o.Dir, bundleName)); err != nil {
					t.Fatal(err)
				}
			case "extra":
				writeFile(t, filepath.Join(f.o.Dir, "extra"), []byte("extra"))
			case "manifest":
				writeFile(t, filepath.Join(f.o.Dir, manifestName), []byte(`{"schema":1,"unknown":true}`))
			case "tag":
				f.o.Tag = "v0.2.1"
			case "sha":
				f.o.SHA = strings.Repeat("b", 40)
			case "signature":
				f.signatureError = errors.New("bad certificate")
			case "symlink":
				p := filepath.Join(f.o.Dir, "gz.rb")
				if err := os.Remove(p); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(f.o.Cosign, p); err != nil {
					t.Fatal(err)
				}
			}
			if _, _, err := verify(context.Background(), f.o, f.run); err == nil {
				t.Fatal("invalid signed stage accepted")
			}
		})
	}
}

func TestSnapshotNeverVerifiesAndExistingStagePreserved(t *testing.T) {
	f := newFixture(t)
	f.o.Snapshot = true
	writeJSON(t, filepath.Join(f.o.Dir, "metadata.json"), metadata{Project: "gzh-cli", Tag: f.o.Tag, Version: "0.2.0-SNAPSHOT", Commit: testSHA})
	// snapshot에서도 실제 version을 사용하는 생성 cask를 검증한다.
	b, err := os.ReadFile(filepath.Join(f.o.Dir, "gz.rb"))
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.ReplaceAll(b, []byte(`version "0.2.0"`), []byte(`version "0.2.0-SNAPSHOT"`))
	b = bytes.ReplaceAll(b, []byte("v#{version}"), []byte("v0.2.0"))
	writeFile(t, filepath.Join(f.o.Dir, "gz.rb"), b)
	if err := prepare(context.Background(), f.o, f.run); err != nil {
		t.Fatal(err)
	}
	if err := prepare(context.Background(), f.o, f.run); err == nil {
		t.Fatal("existing stage overwritten")
	}
	f.o.Dir = filepath.Join(f.o.Dir, "publish")
	f.o.Snapshot = false
	writeFile(t, filepath.Join(f.o.Dir, bundleName), []byte("bundle"))
	if _, _, err := verify(context.Background(), f.o, f.run); err == nil {
		t.Fatal("snapshot verified")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fakeAPI(t *testing.T, fail string) (apiClient, *[]string) {
	t.Helper()
	calls := new([]string)
	api := apiClient{getToken: func(tap bool) string {
		if tap {
			return "tap-secret"
		}
		return "release-secret"
	}}
	api.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		*calls = append(*calls, r.Method+" "+r.URL.Host+r.URL.Path)
		want := "Bearer release-secret"
		if strings.Contains(r.URL.Path, "homebrew-tap") {
			want = "Bearer tap-secret"
		}
		if r.Header.Get("Authorization") != want {
			t.Error("credential crossed repository boundary")
		}
		status := 200
		var response any
		switch {
		case strings.Contains(r.URL.Path, "/releases/tags/"):
			status = 404
			response = map[string]string{"message": "Not Found"}
			if fail == "existing" {
				status = 200
				response = githubRelease{ID: 1, Tag: "v0.2.0"}
			}
			if fail == "forbidden" {
				status = 403
			}
			if fail == "bad404" {
				response = map[string]string{"message": "another failure"}
			}
			if fail == "network" {
				return nil, errors.New("transport may contain release-secret, must be redacted")
			}
		case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/releases"):
			response = []githubRelease{}
			if fail == "list-null" {
				response = nil
			}
			if fail == "list-malformed" {
				response = []map[string]any{{}}
			}
			if fail == "list-missing-draft" {
				response = []map[string]any{{"id": 1, "tag_name": "v0.1.0", "prerelease": false}}
			}
			if fail == "draft" {
				response = []githubRelease{{ID: 1, Tag: "v0.2.0", Draft: true}}
			}
		case r.Method == "GET" && strings.Contains(r.URL.Path, "/contents/"):
			response = map[string]string{"sha": strings.Repeat("a", 40), "path": "Casks/gz.rb", "type": "file"}
			if fail == "create" || strings.HasPrefix(fail, "create-conflict") {
				status = 404
				response = map[string]string{"message": "Not Found"}
			}
			if fail == "tap403" {
				status = 403
			}
			if fail == "tapbad404" {
				status = 404
				response = map[string]string{"unexpected": "Not Found"}
			}
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/releases"):
			status = 201
			response = githubRelease{ID: 42, Tag: "v0.2.0", Draft: true}
			if fail == "create-missing-draft" {
				response = map[string]any{"id": 42, "tag_name": "v0.2.0", "prerelease": false}
			}
		case r.URL.Host == "uploads.github.com":
			b, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			status = 201
			sha := "sha256:" + digest(b)
			if fail == "upload" {
				sha = "wrong"
			}
			response = map[string]any{"name": r.URL.Query().Get("name"), "size": len(b), "digest": sha, "state": "uploaded"}
		case r.Method == "PUT":
			if fail == "tap" || strings.HasPrefix(fail, "create-conflict") {
				status = 409
				if fail == "create-conflict422" {
					status = 422
				}
				response = map[string]string{"message": "redacted"}
				break
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			wantSHA := strings.Repeat("a", 40)
			if fail == "create" {
				wantSHA = ""
				status = 201
				if _, exists := body["sha"]; exists {
					t.Error("create must omit sha")
				}
			}
			if body["branch"] != "main" || body["sha"] != wantSHA {
				t.Error("missing tapCAS")
			}
			decoded, err := base64.StdEncoding.DecodeString(body["content"])
			if err != nil || len(decoded) == 0 {
				t.Error("invalid cask payload")
			}
			response = map[string]any{"commit": map[string]string{"sha": strings.Repeat("b", 40)}, "content": map[string]string{"sha": gitBlobSHA(decoded), "path": "Casks/gz.rb"}}
		case r.Method == "PATCH":
			response = githubRelease{ID: 42, Tag: "v0.2.0"}
			if fail == "publish-missing-draft" {
				response = map[string]any{"id": 42, "tag_name": "v0.2.0", "prerelease": false}
			}
			if fail == "publish" {
				status = 500
			}
		default:
			t.Errorf("unexpected HTTP %s %s", r.Method, r.URL)
			status = 500
		}
		b, err := json.Marshal(response)
		if err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewReader(b)), Header: make(http.Header)}, nil
	})}
	return api, calls
}

func TestPublisherFailClosedAndTapLast(t *testing.T) {
	for _, failure := range []string{"", "create", "existing", "draft", "forbidden", "bad404", "network", "tap403", "tapbad404", "list-null", "list-malformed", "list-missing-draft", "create-missing-draft", "publish-missing-draft", "upload", "publish", "tap", "create-conflict", "create-conflict422"} {
		t.Run(failure, func(t *testing.T) {
			f := newFixture(t)
			f.stage(t)
			files, m, err := verify(context.Background(), f.o, f.run)
			if err != nil {
				t.Fatal(err)
			}
			api, calls := fakeAPI(t, failure)
			err = publish(context.Background(), f.o, m, files, f.run, api)
			writes := 0
			tap, published := false, false
			for _, c := range *calls {
				if !strings.HasPrefix(c, "GET ") {
					writes++
				}
				tap = tap || strings.HasPrefix(c, "PUT ")
				published = published || strings.HasPrefix(c, "PATCH ")
			}
			if failure == "" || failure == "create" {
				if err != nil || !published || !strings.HasPrefix((*calls)[len(*calls)-1], "PUT ") {
					t.Fatalf("publish order/result: %v %v", err, *calls)
				}
			} else {
				if err == nil {
					t.Fatal("failure accepted")
				}
				if (failure == "tap" || strings.HasPrefix(failure, "create-conflict")) && (!published || !strings.Contains(err.Error(), "valid public release retained")) {
					t.Fatal("public release stage lost after tap failure")
				}
			}
			if (failure == "existing" || failure == "draft" || failure == "forbidden" || failure == "bad404" || failure == "network" || failure == "tap403" || failure == "tapbad404") && writes != 0 {
				t.Fatal("existing release mutated")
			}
			if strings.HasPrefix(failure, "list-") && writes != 0 {
				t.Fatal("malformed release listing allowed writes")
			}
			if (failure == "upload" || failure == "publish" || failure == "publish-missing-draft" || failure == "create-missing-draft") && tap {
				t.Fatal("tap changed before successful public release")
			}
			if failure == "upload" && published {
				t.Fatal("published before verified uploads")
			}
			if tap {
				patchIndex, putIndex := -1, -1
				for i, c := range *calls {
					if strings.HasPrefix(c, "PATCH ") {
						patchIndex = i
					}
					if strings.HasPrefix(c, "PUT ") {
						putIndex = i
					}
				}
				if patchIndex < 0 || putIndex <= patchIndex {
					t.Fatal("tap write preceded public assets")
				}
			}
			if err != nil && (strings.Contains(err.Error(), "tap-secret") || strings.Contains(err.Error(), "release-secret")) {
				t.Fatal("credential leaked")
			}
		})
	}
}

func TestGitBlobAndPathLimits(t *testing.T) {
	if gitBlobSHA(nil) != "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391" {
		t.Fatal("Git empty blob identity mismatch")
	}
	for _, name := range []string{"../escape", "a/b", "a\\b", "a b", "bad\nname", "/absolute", ".."} {
		if validName(name) {
			t.Fatalf("unsafe name accepted: %q", name)
		}
	}
	f := newFixture(t)
	if _, err := regularBytes(f.o.Cosign, 1); err == nil {
		t.Fatal("oversized file accepted")
	}
	if _, err := regularBytes(f.o.Repo, maxFile); err == nil {
		t.Fatal("directory accepted as file")
	}
	var manifest releaseManifest
	if err := decodeStrict([]byte(`{} {}`), &manifest); err == nil {
		t.Fatal("multiple JSON values accepted")
	}
}

func TestSignatureVerificationCannotSwapStage(t *testing.T) {
	f := newFixture(t)
	f.stage(t)
	run := func(ctx context.Context, executable string, args, env []string, cwd string) ([]byte, error) {
		if executable == f.o.Cosign && args[0] == "verify-blob" {
			writeFile(t, filepath.Join(f.o.Dir, "gz.rb"), []byte("changed during verification"))
		}
		return f.run(ctx, executable, args, env, cwd)
	}
	if _, _, err := verify(context.Background(), f.o, run); err == nil {
		t.Fatal("verified mutated stage")
	}
}

func TestSignedASwapRestoreCannotAuthenticateUnsignedB(t *testing.T) {
	f := newFixture(t)
	f.stage(t)
	signedA, err := readStage(f.o.Dir)
	if err != nil {
		t.Fatal(err)
	}
	var m releaseManifest
	if err := decodeStrict(signedA[manifestName], &m); err != nil {
		t.Fatal(err)
	}
	unsignedB := make(map[string][]byte, len(signedA))
	for n, b := range signedA {
		unsignedB[n] = append([]byte(nil), b...)
	}
	unsignedB["release-notes.md"] = []byte("unsigned substituted release notes")
	for i := range m.Assets {
		if m.Assets[i].Name == "release-notes.md" {
			m.Assets[i].SHA256 = digest(unsignedB["release-notes.md"])
			m.Assets[i].Size = int64(len(unsignedB["release-notes.md"]))
		}
	}
	manifest, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	unsignedB[manifestName] = manifest
	checksummed := make(map[string][]byte)
	for n, b := range unsignedB {
		if n != checksumName && n != bundleName {
			checksummed[n] = b
		}
	}
	unsignedB[checksumName] = makeChecksums(checksummed)
	for n, b := range unsignedB {
		writeFile(t, filepath.Join(f.o.Dir, n), b)
	}
	var private string
	verifier := func(ctx context.Context, executable string, args, env []string, cwd string) ([]byte, error) {
		if executable != f.o.Cosign || args[0] != "verify-blob" {
			return f.run(ctx, executable, args, env, cwd)
		}
		private = filepath.Dir(args[len(args)-1])
		if private == f.o.Dir {
			t.Fatal("verifier still uses original mutable paths")
		}
		// 공격자가 검증 시점에 signed A를 노출하고 검사 직후 B를 복원한다.
		for n, b := range signedA {
			writeFile(t, filepath.Join(f.o.Dir, n), b)
		}
		defer func() {
			for n, b := range unsignedB {
				writeFile(t, filepath.Join(f.o.Dir, n), b)
			}
		}()
		actual, err := os.ReadFile(args[len(args)-1])
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(actual, unsignedB[checksumName]) {
			t.Fatal("private snapshot differs from publisher memory")
		}
		if bytes.Equal(actual, signedA[checksumName]) {
			return nil, nil
		}
		return nil, errors.New("signed A bundle does not authenticate B checksum bytes")
	}
	if _, _, err := verify(context.Background(), f.o, verifier); err == nil || !strings.Contains(err.Error(), "signature verification failed") {
		t.Fatalf("unsigned B accepted: %v", err)
	}
	if private == "" {
		t.Fatal("signature verification was not exercised")
	}
	if _, err := os.Stat(private); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("failed verification left private snapshot")
	}
}

func TestCLIRejectsInvalidOperationsBeforeSideEffects(t *testing.T) {
	for _, args := range [][]string{nil, {"unknown"}, {"prepare", "--invalid"}, {"verify", "unexpected"}, {"verify", "--snapshot"}, {"publish", "--sha", testSHA, "--tag", "v0.2.0", "--dir", filepath.Join(t.TempDir(), "absent")}, {"prepare", "--sha", "bad", "--tag", "v0.2.0"}} {
		if err := runCLI(args, io.Discard); err == nil {
			t.Fatalf("invalid CLI accepted: %v", args)
		}
	}
}

func TestReleaseArtifactCommandHelper(t *testing.T) {
	t.Helper()
	if os.Getenv("RELEASE_ARTIFACT_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg != "--" || i+1 >= len(os.Args) {
			continue
		}
		switch os.Args[i+1] {
		case "ok":
			fmt.Fprint(os.Stdout, "controlled output")
			os.Exit(0)
		case "fail":
			fmt.Fprint(os.Stderr, "controlled failure")
			os.Exit(7)
		case "large":
			fmt.Fprint(os.Stdout, strings.Repeat("x", 100<<10))
			os.Exit(0)
		}
	}
	os.Exit(9)
}

func TestCommandRunnerBoundsAndCancellation(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	env := append(cleanEnvironment(), "RELEASE_ARTIFACT_HELPER=1")
	for _, mode := range []string{"ok", "fail", "large"} {
		b, err := execute(context.Background(), executable, []string{"-test.run=^TestReleaseArtifactCommandHelper$", "--", mode}, env, "")
		if mode == "ok" {
			if err != nil || string(b) != "controlled output" {
				t.Fatalf("helper output %q %v", b, err)
			}
		} else if err == nil {
			t.Fatalf("%s accepted", mode)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := execute(ctx, executable, nil, env, ""); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel cause lost: %v", err)
	}
	if _, err := execute(context.Background(), filepath.Join(t.TempDir(), "missing"), nil, env, ""); err == nil {
		t.Fatal("missing command accepted")
	}
	var limited limitedBuffer
	if _, err := limited.Write(make([]byte, 65<<10)); err == nil {
		t.Fatal("unbounded output accepted")
	}
}

func TestProductionAPIRejectsRedirectAndBoundsResponse(t *testing.T) {
	api := productionAPI()
	api.getToken = func(bool) string { return "test-secret" }
	calls := 0
	api.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 302, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{"Location": []string{"https://untrusted.invalid/token"}}}, nil
	})
	code, _, err := api.request(context.Background(), false, "GET", "api.github.com", "/repos/Gizzahub/gzh-cli", nil, "application/json")
	if err != nil || code != 302 || calls != 1 {
		t.Fatalf("redirect followed: %d %d %v", code, calls, err)
	}
	if _, _, err := api.request(context.Background(), false, "GET", "evil.invalid", "/", nil, ""); err == nil {
		t.Fatal("untrusted host accepted")
	}
	if _, _, err := api.request(context.Background(), false, "GET", "api.github.com", "//evil.invalid", nil, ""); err == nil {
		t.Fatal("network path accepted")
	}
	api.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", maxJSON+1))), Header: make(http.Header)}, nil
	})
	if _, _, err := api.request(context.Background(), false, "GET", "api.github.com", "/", nil, ""); err == nil {
		t.Fatal("unbounded response accepted")
	}
}
