#!/bin/sh
set -eu

docker compose build --progress plain
exec docker compose up --no-build "$@"
