// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package scan

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli/internal/idecore"
)

// 실제 기본 factory는 격리된 cache를 읽고 취소를 전달해야 한다.
func TestDefaultDetectorCachedFixture(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("PATH", t.TempDir())
	updated := time.Date(2026, 10, 1, 2, 3, 4, 0, time.UTC)
	fixture := idecore.IDE{Name: "Fixture Editor", Type: "fixture-type", Executable: filepath.Join(home, "fixture-editor"), Version: "fixture-9.8.7", Aliases: []string{"fixture-alias"}, InstallMethod: "fixture-method", InstallPath: filepath.Join(home, "fixture-install"), LastUpdated: updated}
	path := filepath.Join(home, ".gz", "cache", "ide.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	data, err := json.Marshal(idecore.IDECache{Timestamp: time.Now(), IDEs: []idecore.IDE{fixture}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
	detector := NewIDEDetector()
	got, err := detector.DetectIDEs(context.Background(), true)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, fixture.Name, got[0].Name)
	require.Equal(t, fixture.Type, got[0].Type)
	require.Equal(t, fixture.Executable, got[0].Executable)
	require.Equal(t, fixture.Version, got[0].Version)
	require.Equal(t, fixture.Aliases, got[0].Aliases)
	require.Empty(t, got[0].ProductCode, "backend에 없는 product code를 만들면 안 된다")

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, data, after, "cache hit must not refresh or rewrite source")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = detector.DetectIDEs(ctx, false)
	require.ErrorIs(t, err, context.Canceled)
	empty, err := json.Marshal(idecore.IDECache{Timestamp: time.Now(), IDEs: []idecore.IDE{}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, empty, 0o600))
	got, err = detector.DetectIDEs(context.Background(), true)
	require.NoError(t, err)
	require.Empty(t, got, "empty real cache must not fall back to fabricated IDEs")
}

// 기존 backend 인터페이스로 전달 계약을 확인하며 감지·프로세스 실행은 하지 않는다.
type forwardingDetector struct {
	context  context.Context
	useCache bool
	calls    int
	err      error
}

func (d *forwardingDetector) DetectIDEs(ctx context.Context, useCache bool) ([]idecore.IDE, error) {
	d.context, d.useCache = ctx, useCache
	d.calls++
	return []idecore.IDE{{Name: "must-not-escape-on-error"}}, d.err
}

func (d *forwardingDetector) FindIDEByAlias([]idecore.IDE, string) *idecore.IDE {
	return nil
}

func TestDetectorForwardsContextCacheAndError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sentinel := errors.New("fixture detection failure")
	for _, useCache := range []bool{true, false} {
		backend := &forwardingDetector{err: sentinel}
		adapter := &coreDetector{core: backend}
		got, err := adapter.DetectIDEs(ctx, useCache)
		require.True(t, backend.context == ctx, "exact caller context must reach backend")
		require.Equal(t, useCache, backend.useCache)
		require.Equal(t, 1, backend.calls)
		require.Same(t, sentinel, err, "backend error identity must be preserved")
		require.Nil(t, got, "partial backend result must not become a successful DTO")
	}
}
