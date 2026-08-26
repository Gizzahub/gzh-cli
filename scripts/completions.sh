#!/bin/sh
set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run ./scripts/completiongen "$sh" >"completions/gzh-manager.$sh"
done
