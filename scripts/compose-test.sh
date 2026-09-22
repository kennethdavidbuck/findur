#!/bin/sh
set -eu

cleanup() {
  docker compose --profile test rm --stop --force browser integration >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM

docker compose build
docker compose up --no-build --wait --force-recreate wiremock
docker compose up --no-build --wait
docker compose --profile test run --rm integration
docker compose --profile test run --rm integration-go
