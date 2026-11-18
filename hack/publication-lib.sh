#!/usr/bin/env bash
# Shared helpers for the publication scripts. Source, don't execute.
# `crane` is taken from $CRANE (default: crane on PATH).

CRANE="${CRANE:-crane}"

die() {
  echo "ERROR: $*" >&2
  exit 1
}

# digest_or_absent <ref>
# Prints the digest and returns 0 if the ref exists, 1 if it is absent (MANIFEST_UNKNOWN /
# NAME_UNKNOWN / 404). Any other failure (403, auth, network) returns 2, so permission problems are
# never mistaken for "not there". Callers must tell 1 and 2 apart:
#   rc=0; have="$(digest_or_absent "$ref")" || rc=$?
digest_or_absent() {
  local out err rc=0
  err="$(mktemp)"
  out="$("$CRANE" digest "$1" 2>"$err")" || rc=$?
  if [[ $rc -eq 0 ]]; then
    rm -f "$err"
    echo "$out"
    return 0
  fi
  if grep -qiE 'MANIFEST_UNKNOWN|NAME_UNKNOWN|status code 404|404 Not Found' "$err"; then
    rm -f "$err"
    return 1
  fi
  echo "ERROR: reading $1 failed (not a plain 'not found'):" >&2
  cat "$err" >&2
  rm -f "$err"
  return 2
}

# child_manifests <name> <index_digest>
# Prints "<os>/<arch>[/<variant>]=<digest>" for each child of an index, sorted.
# Returns 1 if the manifest is not an image index, 2 if it cannot be read.
child_manifests() {
  local m
  m="$("$CRANE" manifest "$1@$2")" || return 2
  jq -e '.manifests | type == "array"' <<<"$m" >/dev/null 2>&1 || return 1
  jq -r '.manifests[] | "\(.platform.os // "unknown")/\(.platform.architecture // "unknown")\(if .platform.variant then "/" + .platform.variant else "" end)=\(.digest)"' <<<"$m" | sort
}

# Publication record written by publish-ecr.sh, one image per line:
#   <platform|index> <tag> <sha256:digest>
# e.g. "linux/amd64 <sha>-amd64 sha256:..." and "index <sha> sha256:...".
