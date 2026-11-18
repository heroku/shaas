#!/usr/bin/env bash
# Build the platform images once, push them to ECR (the source registry) and assemble the
# multi-platform index there. Writes a publication record for publish-rr.sh and
# verify-publication.sh.
#
# Usage: publish-ecr.sh <ecr repo> <tag> "<arch> [<arch>...]" <record file to write>
#   e.g. publish-ecr.sh 970597968373.dkr.ecr.us-east-1.amazonaws.com/s/heroku/shaas/shaas $GITHUB_SHA "amd64 arm64" tmp/publication/record.txt
#
# Platform images are tagged <tag>-<arch>; the index is tagged <tag>. The index is made from the
# pushed images by digest (no rebuild, no registry-side re-push of layers).
#
# Safe to re-run for the same <tag> (ECR repos are immutable): a platform image or index that
# already exists is reused, never rebuilt or overwritten; an existing index whose children differ
# from the platform images is a hard failure.
#
# Env: DOCKER (default: docker), CRANE (default: crane).
set -euo pipefail
# shellcheck source=hack/publication-lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/publication-lib.sh"
DOCKER="${DOCKER:-docker}"

[[ $# -eq 4 ]] || die "usage: $0 <ecr repo> <tag> \"<arch>...\" <record file to write>"
repo="${1%/}" tag="$2" archs="$3" record="$4"
[[ -n "${archs// /}" ]] || die "no architectures given (is the ARCHITECTURES secret set?)"
[[ "$tag" != *-dirty* ]] || die "refusing to publish dirty tag $tag"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
lines=() children=() want_kids=()

for arch in $archs; do
  [[ "$arch" =~ ^[a-z0-9]+$ ]] || die "invalid architecture: $arch"
  arch_tag="$tag-$arch" rc=0
  digest="$(digest_or_absent "$repo:$arch_tag")" || rc=$?
  case $rc in
    0) echo "  $repo:$arch_tag already present ($digest); reusing, not rebuilding" ;;
    1)
      echo "Building and pushing linux/$arch"
      # --provenance/--sbom=false: the image has never carried attestations; an attestation
      # manifest would add an extra, platform-less child to the index.
      "$DOCKER" buildx build --platform "linux/$arch" --build-arg "ARCH=$arch" -f Dockerfile \
        --provenance=false --sbom=false --push --metadata-file "$work/$arch.json" -t "$repo:$arch_tag" .
      digest="$(jq -r '."containerimage.digest" // empty' "$work/$arch.json")"
      [[ "$digest" == sha256:* ]] || die "build of linux/$arch did not report an image digest"
      got="$(digest_or_absent "$repo:$arch_tag")" || die "cannot read back $repo:$arch_tag after push"
      [[ "$got" == "$digest" ]] || die "$repo:$arch_tag is $got but the build produced $digest"
      ;;
    *) die "cannot read $repo:$arch_tag" ;;
  esac
  lines+=("linux/$arch $arch_tag $digest")
  children+=("$repo@$digest")
  want_kids+=("linux/$arch=$digest")
done

expected="$(printf '%s\n' "${want_kids[@]}" | sort)"
rc=0
index="$(digest_or_absent "$repo:$tag")" || rc=$?
case $rc in
  0)
    actual="$(child_manifests "$repo" "$index")" || die "$repo:$tag exists but is not a readable image index"
    [[ "$actual" == "$expected" ]] || die "$repo:$tag already exists with different children (immutable tag); expected: $(tr '\n' ' ' <<<"$expected") got: $(tr '\n' ' ' <<<"$actual")"
    echo "  $repo:$tag already present ($index) with the expected children"
    ;;
  1)
    echo "Assembling index $repo:$tag"
    "$DOCKER" buildx imagetools create --tag "$repo:$tag" "${children[@]}"
    index="$(digest_or_absent "$repo:$tag")" || die "cannot read back $repo:$tag after creating the index"
    ;;
  *) die "cannot read $repo:$tag" ;;
esac

mkdir -p "$(dirname "$record")"
{
  printf '%s\n' "${lines[@]}"
  echo "index $tag $index"
} >"$record"
echo "Wrote $record"
