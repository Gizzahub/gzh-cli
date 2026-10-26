// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// releaseartifact는 서명한 CI 산출물만 검증하고 게시한다.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	if err := runCLI(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "releaseartifact:", err)
		os.Exit(1)
	}
}

func runCLI(args []string, output io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("expected prepare, verify, or publish")
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(output)
	dir := fs.String("dir", "dist", "GoReleaser dist or prepared publish directory")
	tag := fs.String("tag", "", "exact stable vMAJOR.MINOR.PATCH tag")
	sha := fs.String("sha", "", "exact source commit SHA")
	repo := fs.String("repo", ".", "local source Git repository")
	cosign := fs.String("cosign", "", "absolute pinned cosign v3.1.3 executable")
	snapshot := fs.Bool("snapshot", false, "prepare explicitly unpublishable snapshot")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	o := options{Dir: *dir, Tag: *tag, SHA: *sha, Repo: *repo, Cosign: *cosign, Snapshot: *snapshot}
	switch args[0] {
	case "prepare":
		return prepare(ctx, o, execute)
	case "verify", "publish":
		if o.Snapshot {
			return fmt.Errorf("snapshot cannot be verified or published")
		}
		files, manifest, err := verify(ctx, o, execute)
		if err != nil {
			return err
		}
		if args[0] == "verify" {
			return nil
		}
		return publish(ctx, o, manifest, files, execute, productionAPI())
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
