// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SimpleSSHInstaller uses system SSH command for installation.
type SimpleSSHInstaller struct{}

// NewSimpleSSHInstaller creates a new simple SSH installer.
func NewSimpleSSHInstaller() *SimpleSSHInstaller {
	return &SimpleSSHInstaller{}
}

// InstallPublicKeySimple installs a public key using system SSH.
//
// ctx는 실행 맥락이다. 이 꾸러미의 다른 바깥 명령 호출(gcloud, az, aws,
// kubectl, docker)은 전부 맥락을 넘기는데 여기만 context.Background()였다.
func (installer *SimpleSSHInstaller) InstallPublicKeySimple(ctx context.Context, host, user, publicKeyPath string) error {
	return installer.InstallPublicKeySimpleWithOptions(ctx, host, user, publicKeyPath, HostKeyOptions{})
}

// InstallPublicKeySimpleWithOptions installs a public key with explicit host-key verification options.
func (installer *SimpleSSHInstaller) InstallPublicKeySimpleWithOptions(
	ctx context.Context,
	host, user, publicKeyPath string,
	hostKeyOptions HostKeyOptions,
) error {
	// Read public key
	keyContent, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	key := strings.TrimSpace(string(keyContent))
	if key == "" {
		return fmt.Errorf("public key file is empty")
	}

	// Basic validation
	if !strings.HasPrefix(key, "ssh-") {
		return fmt.Errorf("invalid public key format")
	}

	fmt.Printf("Installing public key to %s@%s...\n", user, host)

	// Use ssh to install the key
	commands := []string{
		"mkdir -p ~/.ssh",
		"chmod 700 ~/.ssh",
		fmt.Sprintf("echo '%s' >> ~/.ssh/authorized_keys", key),
		"chmod 600 ~/.ssh/authorized_keys",
		// Remove duplicates
		"sort ~/.ssh/authorized_keys | uniq > ~/.ssh/authorized_keys.tmp && mv ~/.ssh/authorized_keys.tmp ~/.ssh/authorized_keys",
		"echo 'Public key installed successfully'",
	}

	cmdStr := strings.Join(commands, " && ")
	knownHostsPath, err := resolveKnownHostsPath(hostKeyOptions.KnownHostsPath)
	if err != nil {
		return err
	}
	if err := prepareKnownHostsFile(knownHostsPath, hostKeyOptions.AcceptNewHostKey); err != nil {
		return err
	}
	// 응답 없는 호스트에 매달리지 않도록 붙는 시간을 제한한다. ssh는 기본값이
	// 없으면 TCP가 포기할 때까지 기다린다. 맥락 취소는 이미 붙은 뒤에 끊는
	// 길이고, ConnectTimeout은 붙기 전에 포기하는 길이라 둘 다 필요하다.
	cmd := exec.CommandContext(ctx, "ssh", buildSimpleSSHArgs(host, user, cmdStr, knownHostsPath, hostKeyOptions.AcceptNewHostKey)...)

	// Connect stdin/stdout/stderr to allow password input
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func buildSimpleSSHArgs(host, user, command, knownHostsPath string, acceptNew bool) []string {
	strictHostKeyChecking := "yes"
	if acceptNew {
		strictHostKeyChecking = "accept-new"
	}

	return []string{
		"-o", "StrictHostKeyChecking=" + strictHostKeyChecking,
		"-o", "UserKnownHostsFile=" + knownHostsPath,
		"-o", "GlobalKnownHostsFile=" + os.DevNull,
		"-o", "KnownHostsCommand=none",
		"-o", "VerifyHostKeyDNS=no",
		"-o", "UpdateHostKeys=no",
		"-o", "CheckHostIP=no",
		"-o", "HashKnownHosts=no",
		"-o", "ConnectTimeout=10",
		fmt.Sprintf("%s@%s", user, host), command,
	}
}
