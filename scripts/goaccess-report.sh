#!/usr/bin/env bash
set -euo pipefail

LOG_DIR="${LOG_DIR:-/var/log/welink/nginx}"
OUT_DIR="${OUT_DIR:-/var/log/welink/goaccess}"
IMAGE="${GOACCESS_IMAGE:-allinurl/goaccess}"
ACCESS_LOG="$LOG_DIR/access.log"

LOG_FORMAT='%v %h %^ %^ [%d:%t %^] "%r" %s %b "%R" "%u"'
DATE_FORMAT='%d/%b/%Y'
TIME_FORMAT='%H:%M:%S'

if [ ! -s "$ACCESS_LOG" ]; then
  echo "access log not found or empty: $ACCESS_LOG" >&2
  exit 1
fi

mkdir -p "$OUT_DIR/welink" "$OUT_DIR/demo"

run_goaccess() {
  local output="$1"
  docker run --rm -i \
    -v "$OUT_DIR:/report" \
    "$IMAGE" - \
    --log-format="$LOG_FORMAT" \
    --date-format="$DATE_FORMAT" \
    --time-format="$TIME_FORMAT" \
    --ignore-panel=REQUESTS_STATIC \
    -o "$output"
}

if grep -Eq '^(welink\.click|www\.welink\.click) ' "$ACCESS_LOG"; then
  grep -E '^(welink\.click|www\.welink\.click) ' "$ACCESS_LOG" \
    | run_goaccess /report/welink/index.html
  echo "generated $OUT_DIR/welink/index.html"
else
  echo "no welink.click visits found yet"
fi

if grep -q '^demo\.welink\.click ' "$ACCESS_LOG"; then
  grep '^demo\.welink\.click ' "$ACCESS_LOG" \
    | run_goaccess /report/demo/index.html
  echo "generated $OUT_DIR/demo/index.html"
else
  echo "no demo.welink.click visits found yet"
fi
