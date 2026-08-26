#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
SOURCE_DATE_EPOCH="$(git log -1 --format=%ct)" \
	go run ./scripts/manpagegen >manpages/gzh-manager.1.gz
