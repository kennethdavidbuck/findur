#!/bin/sh
set -eu

exec docker compose logs --follow --tail 100 "$@"
