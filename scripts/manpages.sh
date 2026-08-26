#!/bin/sh
set -eu

mkdir -p manpages
SOURCE_DATE_EPOCH="$(git log -1 --format=%ct)" \
	go run ./scripts/manpagegen manpages/gzh-manager.1.gz
