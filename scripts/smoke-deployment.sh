#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  echo "usage: $0 https://your-static-site.onrender.com" >&2
  exit 2
fi

if ! printf '%s\n' "$1" | grep -Eq '^https://[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?)*(:[1-9][0-9]{0,4})?/?$'; then
  echo "deployment URL must be an HTTPS origin without a path, query, fragment, or credentials" >&2
  exit 2
fi

base_url=${1%/}
authority=${base_url#https://}
case "$authority" in
  *:*)
    port=${authority##*:}
    if [ "$port" -gt 65535 ]; then
      echo "deployment URL port must be from 1 through 65535" >&2
      exit 2
    fi
    ;;
esac

smoke_dir=$(mktemp -d)
trap 'rm -rf "$smoke_dir"' EXIT HUP INT TERM

curl --fail --silent --show-error --max-time 15 "$base_url/" --output "$smoke_dir/root.html"
grep -q '<div id="root"></div>' "$smoke_dir/root.html"

module_path=$(sed -n 's/.*<script[^>]*type="module"[^>]*src="\([^"]*\)"[^>]*>.*/\1/p' "$smoke_dir/root.html")
case "$module_path" in
  /?*) ;;
  *)
    echo "deployed shell does not reference a same-origin module asset" >&2
    exit 1
    ;;
esac

module_type=$(curl --fail --silent --show-error --max-time 15 "$base_url$module_path" \
  --output "$smoke_dir/module.js" --write-out '%{content_type}')
case "$module_type" in
  text/html*|application/xhtml+xml*)
    echo "module asset resolved to an HTML fallback" >&2
    exit 1
    ;;
esac
if [ ! -s "$smoke_dir/module.js" ] || grep -Eiq '<(!doctype|html|body)([[:space:]>])' "$smoke_dir/module.js"; then
  echo "module asset is empty or contains an HTML fallback" >&2
  exit 1
fi

curl --fail --silent --show-error --max-time 15 "$base_url/api/healthz" --output "$smoke_dir/health.json"
grep -Eq '^\{"status":"ok"\}[[:space:]]*$' "$smoke_dir/health.json"

curl --fail --silent --show-error --max-time 15 "$base_url/api/readyz" --output "$smoke_dir/ready.json"
grep -Eq '^\{"status":"ready"\}[[:space:]]*$' "$smoke_dir/ready.json"

echo "deployment smoke check passed"
