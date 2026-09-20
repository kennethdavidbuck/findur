#!/bin/sh
set -eu

docker compose --profile dev watch backend frontend-dev "$@"
