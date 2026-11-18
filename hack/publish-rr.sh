#!/usr/bin/env bash
# Copy the images published to ECR to the Runtime Registry (RR) with identical content.
#
# RR-ONLY STAGE: delete this script and the RR steps in .github/workflows/docker.yml when RR is
# retired.
#
# Usage: publish-rr.sh <publication record> <ecr repo> <rr repo>
#
# Copies by digest with `crane copy` (index, platform children and their bytes unchanged). Never
# rebuilds and never reads from RR to write to ECR. Safe to re-run: content already present with
# the same digest is accepted; a tag that exists with different content is a hard failure (tags
# are immutable). The existing RR tags (<tag>-<arch> and <tag>) keep working.
set -euo pipefail
# shellcheck source=hack/publication-lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/publication-lib.sh"

[[ $# -eq 3 ]] || die "usage: $0 <publication record> <ecr repo> <rr repo>"
record="$1" src_repo="${2%/}" dst_repo="${3%/}"
[[ -s "$record" ]] || die "$record is missing or empty; publish to ECR first"

while read -r _ tag digest; do
  [[ -n "${tag:-}" ]] || continue
  [[ "$tag" != *-dirty* ]] || die "refusing to publish dirty tag $tag"
  [[ "$digest" == sha256:* ]] || die "bad digest in record for $tag: $digest"
  have="" rc=0
  have="$(digest_or_absent "$dst_repo:$tag")" || rc=$?
  case $rc in
    0)
      [[ "$have" == "$digest" ]] || die "$dst_repo:$tag already exists with $have but ECR has $digest; refusing to overwrite an immutable tag"
      echo "  $dst_repo:$tag already present ($digest)"
      ;;
    1)
      "$CRANE" copy "$src_repo@$digest" "$dst_repo:$tag"
      echo "  copied $dst_repo:$tag ($digest)"
      ;;
    *) die "cannot read $dst_repo:$tag" ;;
  esac
done <"$record"
