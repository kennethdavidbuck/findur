#!/bin/sh
set -eu

docker compose --profile dev --profile test down --remove-orphans "$@"
