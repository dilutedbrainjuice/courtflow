#!/usr/bin/env bash
set -u
missing=0
for tool in git go node npm podman; do
  if command -v "$tool" >/dev/null 2>&1; then
    case "$tool" in
      go) go version ;;
      *) "$tool" --version ;;
    esac
  else
    printf 'Missing: %s\n' "$tool"
    missing=1
  fi
done
if command -v psql >/dev/null 2>&1; then
  psql --version
else
  printf 'PostgreSQL client absent; use a local Podman PostgreSQL container when configured.\n'
fi
exit "$missing"
