#!/bin/bash
# Install as /usr/local/bin/blog-deploy, owned by root and executable.
# Give the deployment key only a forced command invoking this script.
set -euo pipefail
umask 022
ROOT=/data/oneyearago
REPO=https://github.com/changan29/changan29.github.io.git
exec 9>"$ROOT/.deploy.lock"
flock -w 60 9
command=${SSH_ORIGINAL_COMMAND:-}
if [[ ! $command =~ ^deploy\ ([0-9a-f]{40})$ ]]; then
    echo 'Only a site deployment commit is accepted.' >&2
    exit 1
fi
commit=${BASH_REMATCH[1]}
if [ ! -d "$ROOT/site-repo.git" ]; then
    git clone --bare --single-branch --branch blog-site "$REPO" "$ROOT/site-repo.git"
fi
git --git-dir="$ROOT/site-repo.git" fetch --no-tags origin refs/heads/blog-site:refs/heads/blog-site
latest=$(git --git-dir="$ROOT/site-repo.git" rev-parse refs/heads/blog-site)
if [ "$commit" != "$latest" ]; then
    echo 'Requested release is not the latest generated site.' >&2
    exit 1
fi
mkdir -p "$ROOT/releases"
release="$ROOT/releases/$commit"
if [ ! -d "$release" ]; then
    staging=$(mktemp -d "$ROOT/releases/.staging.XXXXXX")
    trap 'rm -rf -- "${staging:-}"' EXIT
    git --git-dir="$ROOT/site-repo.git" archive "$commit" | tar -x -C "$staging"
    test -s "$staging/index.html"
    test -s "$staging/admin/index.html"
    test -s "$staging/archives/index.html"
    test -s "$staging/release.json"
    # Retain pre-existing standalone demos outside the old generated repository.
    if [ -d "$ROOT/legacy-extra" ]; then cp -a "$ROOT/legacy-extra/." "$staging/"; fi
    mv "$staging" "$release"
    trap - EXIT
fi
previous=$(readlink "$ROOT/current" || true)
rm -f "$ROOT/current.next"
ln -s "$release" "$ROOT/current.next"
mv -Tf "$ROOT/current.next" "$ROOT/current"
expected_source=$(python -c 'import json,sys; print(json.load(open(sys.argv[1]))["source_commit"])' "$release/release.json")
if ! curl --fail --silent --show-error --resolve www.oneyearago.me:443:127.0.0.1 https://www.oneyearago.me/release.json | python -c 'import json,sys; release=json.load(sys.stdin); assert release["generator"] == "hugo" and release["source_commit"] == sys.argv[1]' "$expected_source"; then
    if [ -n "$previous" ]; then ln -s "$previous" "$ROOT/current.rollback";mv -Tf "$ROOT/current.rollback" "$ROOT/current"; fi
    echo 'Site health check failed; previous release restored.' >&2
    exit 1
fi
printf 'Deployed %s\n' "$commit"
