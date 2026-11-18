#!/usr/bin/env bash
# Verify that the release is complete and identical in ECR and (optionally) RR. Run it after
# publishing and before anything advertises the release. Fails closed.
#
# Usage: verify-publication.sh <publication record> <ecr repo> [<rr repo>]
#   Leave <rr repo> empty (or omit it) to check ECR only; keep that form once RR is retired.
#
# Checks, per registry: every recorded tag resolves to the recorded digest and is readable; the
# index has exactly the recorded platforms with the recorded child digests, and every child is
# readable. RR must also have the same children as ECR. Any read error that is not a plain
# "not found" (403, auth, network) is a failure.
set -euo pipefail
# shellcheck source=hack/publication-lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/publication-lib.sh"

[[ $# -ge 2 && $# -le 3 ]] || die "usage: $0 <publication record> <ecr repo> [<rr repo>]"
record="$1" src_repo="${2%/}" dst_repo="${3:-}"
dst_repo="${dst_repo%/}"
OWNER="${PUBLICATION_OWNER:-heroku/runtime (heroku/shaas)}"
REMEDIATION="${PUBLICATION_REMEDIATION:-docs/release-publication.md}"
[[ -s "$record" ]] || die "$record is missing or empty; publish first"

failures=0
fail() { # <image> <expected ref> <reason>
  failures=$((failures + 1))
  printf 'FAIL image=%s expected=%s reason="%s" owner="%s" remediation=%s\n' "$1" "$2" "$3" "$OWNER" "$REMEDIATION" >&2
  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
    echo "::error title=shaas publication check failed::$1 ($2): $3. Owner: $OWNER. See $REMEDIATION"
  fi
}

index_tag="" index_digest="" expected=""
while read -r plat tag digest; do
  [[ -n "${tag:-}" ]] || continue
  [[ "$tag" != *-dirty* ]] || die "record contains dirty tag $tag"
  if [[ "$plat" == index ]]; then
    index_tag="$tag" index_digest="$digest"
  else
    expected+="$plat=$digest"$'\n'
  fi
done <"$record"
[[ -n "$index_tag" && -n "$expected" ]] || die "$record has no index or no platform images"
expected="$(sort <<<"${expected%$'\n'}")"

# check_registry <label> <repo>
check_registry() {
  local label="$1" repo="$2" got rc kids plat kid
  # Every tag in the record resolves to its recorded digest and is readable.
  while read -r plat tag digest; do
    [[ -n "${tag:-}" ]] || continue
    rc=0
    got="$(digest_or_absent "$repo:$tag")" || rc=$?
    if [[ $rc -eq 1 ]]; then
      fail "$repo" "$repo:$tag@$digest" "$label: tag $tag not found"
      continue
    elif [[ $rc -ne 0 ]]; then
      fail "$repo" "$repo:$tag@$digest" "$label: cannot read tag $tag (permission or registry error)"
      continue
    fi
    [[ "$got" == "$digest" ]] || { fail "$repo" "$repo:$tag@$digest" "$label: tag $tag points at $got, not the recorded $digest"; continue; }
    "$CRANE" manifest "$repo@$digest" >/dev/null 2>&1 || fail "$repo" "$repo:$tag@$digest" "$label: manifest $digest is not readable"
  done <"$record"

  # The index has exactly the recorded platforms and children, and each child is readable.
  rc=0
  kids="$(child_manifests "$repo" "$index_digest")" || rc=$?
  if [[ $rc -ne 0 ]]; then
    fail "$repo" "$repo:$index_tag@$index_digest" "$label: not a readable image index"
    return
  fi
  [[ "$kids" == "$expected" ]] || fail "$repo" "$repo:$index_tag@$index_digest" "$label: index children differ from the build (expected $(tr '\n' ' ' <<<"$expected") got $(tr '\n' ' ' <<<"$kids"))"
  while IFS='=' read -r plat kid; do
    [[ -n "$kid" ]] || continue
    "$CRANE" manifest "$repo@$kid" >/dev/null 2>&1 || fail "$repo" "$repo:$index_tag@$index_digest" "$label: platform manifest $plat ($kid) is not readable"
  done <<<"$kids"
}

check_registry "ecr" "$src_repo"
[[ -z "$dst_repo" ]] || check_registry "rr" "$dst_repo"

if [[ $failures -gt 0 ]]; then
  echo "Publication verification failed: $failures problem(s). The release must not be advertised." >&2
  exit 1
fi
echo "Publication verified: $index_tag ($index_digest), platforms: $(cut -d= -f1 <<<"$expected" | tr '\n' ' ')${dst_repo:+ at ecr and rr}."
