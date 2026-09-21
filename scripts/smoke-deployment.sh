#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  echo "usage: $0 https://your-static-site.onrender.com" >&2
  exit 2
fi

if ! printf '%s\n' "$1" | grep -Eq '^(https://[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?)*|http://127\.0\.0\.1)(:[1-9][0-9]{0,4})?/?$'; then
  echo "deployment URL must be an HTTPS origin, or an HTTP loopback origin, without a path, query, fragment, or credentials" >&2
  exit 2
fi

base_url=${1%/}
authority=${base_url#*://}
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

deadline=$(( $(date +%s) + 300 ))
while :; do
  rm -f "$smoke_dir/root.html" "$smoke_dir/status.html" "$smoke_dir/health.json" "$smoke_dir/ready.json"
  curl --fail --silent --show-error --max-time 15 "$base_url/" --output "$smoke_dir/root.html" || true
  curl --fail --silent --show-error --max-time 15 "$base_url/__status" --output "$smoke_dir/status.html" || true
  curl --fail --silent --show-error --max-time 15 "$base_url/api/healthz" --output "$smoke_dir/health.json" || true
  curl --fail --silent --show-error --max-time 15 "$base_url/api/readyz" --output "$smoke_dir/ready.json" || true
  if grep -q '<div id="root"></div>' "$smoke_dir/root.html" 2>/dev/null \
    && grep -q '<div id="root"></div>' "$smoke_dir/status.html" 2>/dev/null \
    && grep -Eq '^\{"status":"ok","buildSha":"[0-9a-f]{40}"\}[[:space:]]*$' "$smoke_dir/health.json" 2>/dev/null \
    && grep -Eq '^\{"status":"ready","buildSha":"[0-9a-f]{40}"\}[[:space:]]*$' "$smoke_dir/ready.json" 2>/dev/null; then
    break
  fi
  if [ "$(date +%s)" -ge "$deadline" ]; then
    echo "deployment did not become healthy" >&2
    exit 1
  fi
  sleep 5
done

grep -q '<div id="root"></div>' "$smoke_dir/root.html"
grep -q '<div id="root"></div>' "$smoke_dir/status.html"
grep -Eq '<meta name="findur-build-sha" content="[0-9a-f]{40}"' "$smoke_dir/root.html"
frontend_sha=$(sed -n 's/.*<meta name="findur-build-sha" content="\([0-9a-f]\{40\}\)".*/\1/p' "$smoke_dir/root.html")
test -n "$frontend_sha"

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
grep -Fq "$frontend_sha" "$smoke_dir/module.js"

grep -Eq '^\{"status":"ok","buildSha":"[0-9a-f]{40}"\}[[:space:]]*$' "$smoke_dir/health.json"

grep -Eq '^\{"status":"ready","buildSha":"[0-9a-f]{40}"\}[[:space:]]*$' "$smoke_dir/ready.json"

echo "deployment health smoke check passed"
