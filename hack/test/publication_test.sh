#!/usr/bin/env bash
# Tests for hack/publish-ecr.sh, hack/publish-rr.sh and hack/verify-publication.sh against a fake
# registry (a directory) with stub `crane` and `docker`. Run: hack/test/publication_test.sh
set -uo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
hack="$here/.."
root="$(mktemp -d)"
trap 'rm -rf "$root"' EXIT
mkdir -p "$root/bin"

# --- fake registry. State: $FAKE/tags/<repo>:<tag> -> digest, $FAKE/manifests/<digest> -> JSON. ---
cat >"$root/bin/crane" <<'STUB'
#!/usr/bin/env bash
set -u
key() { echo "${1//\//_}"; }
deny() { [[ -f "$FAKE/deny" ]] && [[ "$1" == "$(cat "$FAKE/deny")"* ]]; }
case "$1" in
  digest)
    deny "$2" && { echo "denied: 403 Forbidden" >&2; exit 1; }
    f="$FAKE/tags/$(key "$2")"
    [[ -f "$f" ]] || { echo "MANIFEST_UNKNOWN: not found" >&2; exit 1; }
    cat "$f" ;;
  manifest)
    d="${2##*@}"
    [[ -f "$FAKE/manifests/$d" ]] || { echo "not found" >&2; exit 1; }
    cat "$FAKE/manifests/$d" ;;
  copy)
    d="${2##*@}"
    [[ -f "$FAKE/manifests/$d" ]] || { echo "source missing" >&2; exit 1; }
    echo "$d" >"$FAKE/tags/$(key "$3")"
    echo "$3" >>"$FAKE/copies" ;;
  *) echo "unexpected crane $*" >&2; exit 2 ;;
esac
STUB
cat >"$root/bin/docker" <<'STUB'
#!/usr/bin/env bash
set -u
key() { echo "${1//\//_}"; }
next() { local n; n=$(( $(cat "$FAKE/counter" 2>/dev/null || echo 0) + 1 )); echo "$n" >"$FAKE/counter"; printf 'sha256:%064x' "$n"; }
[[ "$1" == buildx ]] || { echo "unexpected docker $*" >&2; exit 2; }
shift
case "$1 $2" in
  "build "*|"build --platform")
    shift; platform="" meta="" tag=""
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --platform) platform="$2"; shift 2 ;;
        --metadata-file) meta="$2"; shift 2 ;;
        -t) tag="$2"; shift 2 ;;
        *) shift ;;
      esac
    done
    echo "$tag" >>"$FAKE/builds"
    [[ -f "$FAKE/fail-build" ]] && exit 1
    d="$(next)"
    mkdir -p "$FAKE/manifests" "$FAKE/tags"
    echo "{\"_platform\":\"$platform\"}" >"$FAKE/manifests/$d"
    echo "$d" >"$FAKE/tags/$(key "$tag")"
    echo "{\"containerimage.digest\":\"$d\"}" >"$meta" ;;
  "imagetools create")
    shift 2; tag=""
    [[ "$1" == --tag ]] && { tag="$2"; shift 2; }
    entries=()
    for c in "$@"; do
      d="${c##*@}"
      p="$(jq -r '._platform' "$FAKE/manifests/$d")"
      entries+=("{\"digest\":\"$d\",\"platform\":{\"os\":\"${p%/*}\",\"architecture\":\"${p#*/}\"}}")
    done
    d="$(next)"
    echo "{\"manifests\":[$(IFS=,; echo "${entries[*]}")]}" >"$FAKE/manifests/$d"
    echo "$d" >"$FAKE/tags/$(key "$tag")"
    echo "$tag" >>"$FAKE/indexes" ;;
  *) echo "unexpected docker buildx $*" >&2; exit 2 ;;
esac
STUB
chmod +x "$root/bin/"*
export PATH="$root/bin:$PATH" CRANE="$root/bin/crane" DOCKER="$root/bin/docker"

pass=0 failed=0
reset() { export FAKE="$root/fake"; rm -rf "$FAKE"; mkdir -p "$FAKE/tags" "$FAKE/manifests"; rec="$root/record.txt"; rm -f "$rec"; }
ECR=ecr.example/s/heroku/shaas/shaas RR=rr.example/s/heroku/shaas SHA=abc123
# expect <name> <expected exit: 0|1> <command...>
expect() {
  local name="$1" want="$2"; shift 2
  local out rc=0
  out="$("$@" 2>&1)" || rc=$?
  if [[ ( $want == 0 && $rc -eq 0 ) || ( $want == 1 && $rc -ne 0 ) ]]; then
    pass=$((pass + 1)); echo "ok   $name"
  else
    failed=$((failed + 1)); echo "FAIL $name (exit $rc, wanted $want)"; printf '     %s\n' "${out//$'\n'/$'\n     '}"
  fi
}
count() { [[ -f "$1" ]] && wc -l <"$1" | tr -d ' ' || echo 0; }
publish() { (cd "$root" && "$hack/publish-ecr.sh" "$ECR" "$SHA" "amd64 arm64" "$rec"); }
to_rr() { "$hack/publish-rr.sh" "$rec" "$ECR" "$RR"; }
verify() { "$hack/verify-publication.sh" "$rec" "$ECR" "$RR"; }
touch "$root/Dockerfile"

reset
expect "publish-ecr builds each platform once and writes a record" 0 publish
[[ "$(count "$FAKE/builds")" == 2 ]] || { failed=$((failed + 1)); echo "FAIL expected exactly 2 builds"; }
expect "publish-rr copies index and platform images" 0 to_rr
[[ "$(count "$FAKE/copies")" == 3 ]] || { failed=$((failed + 1)); echo "FAIL expected 3 copies"; }
expect "verify passes on a complete release" 0 verify

# Retry after a full publication: nothing is rebuilt or recopied.
builds_before="$(count "$FAKE/builds")"
expect "publication-only retry is accepted" 0 bash -c "$(declare -f publish to_rr verify); root='$root' hack='$hack' ECR=$ECR RR=$RR SHA=$SHA rec='$rec'; publish && to_rr && verify"
[[ "$(count "$FAKE/builds")" == "$builds_before" ]] || { failed=$((failed + 1)); echo "FAIL retry rebuilt images"; }
[[ "$(count "$FAKE/copies")" == 3 ]] || { failed=$((failed + 1)); echo "FAIL retry recopied images"; }

# Partial ECR publication: only amd64 pushed before a failure; retry builds only arm64.
reset
publish >/dev/null 2>&1
rm "$FAKE/tags/$(echo "$ECR:$SHA" | tr / _)" "$FAKE/tags/$(echo "$ECR:$SHA-arm64" | tr / _)"
: >"$FAKE/builds"
expect "retry after partial ECR publication rebuilds only the missing platform" 0 publish
[[ "$(count "$FAKE/builds")" == 1 ]] || { failed=$((failed + 1)); echo "FAIL expected 1 rebuild"; }

reset
touch "$FAKE/fail-build"
expect "failed build fails publication" 1 publish
rm "$FAKE/fail-build"

reset
expect "empty architecture list is rejected" 1 "$hack/publish-ecr.sh" "$ECR" "$SHA" " " "$rec"
expect "dirty tag is rejected" 1 "$hack/publish-ecr.sh" "$ECR" "$SHA-dirty" "amd64" "$rec"

# Conflicting immutable index tag in ECR.
reset
publish >/dev/null 2>&1
echo "sha256:$(printf '%064x' 999)" >"$FAKE/tags/$(echo "$ECR:$SHA" | tr / _)"
echo '{"manifests":[]}' >"$FAKE/manifests/sha256:$(printf '%064x' 999)"
expect "existing ECR index with different children is refused" 1 publish

# RR conflicts and verification failures.
reset
publish >/dev/null 2>&1
other="sha256:$(printf '%064x' 998)"
echo "$other" >"$FAKE/tags/$(echo "$RR:$SHA" | tr / _)"
expect "publish-rr refuses a conflicting RR tag" 1 to_rr
echo "$other" >"$FAKE/tags/$(echo "$RR:$SHA" | tr / _)"
expect "verify fails when RR tag has a different digest" 1 verify

reset
publish >/dev/null 2>&1
to_rr >/dev/null 2>&1
rm "$FAKE/tags/$(echo "$RR:$SHA-arm64" | tr / _)"
expect "verify fails when an RR platform tag is missing" 1 verify

reset
publish >/dev/null 2>&1
expect "verify fails when RR is not published yet" 1 verify
expect "verify ECR-only passes without RR" 0 "$hack/verify-publication.sh" "$rec" "$ECR"

reset
publish >/dev/null 2>&1
to_rr >/dev/null 2>&1
d="$(grep ^linux/arm64 "$rec" | cut -d' ' -f3)"
rm "$FAKE/manifests/$d"
expect "verify fails on an unreadable platform manifest" 1 verify

reset
publish >/dev/null 2>&1
to_rr >/dev/null 2>&1
echo "$ECR" >"$FAKE/deny"
expect "verify fails closed on a 403 from ECR" 1 verify
rm "$FAKE/deny"

reset
publish >/dev/null 2>&1
to_rr >/dev/null 2>&1
grep -v '^linux/arm64' "$rec" >"$rec.new" && mv "$rec.new" "$rec"
expect "verify fails when a platform is missing from the index" 1 verify

echo "$pass passed, $failed failed"
[[ $failed -eq 0 ]]
