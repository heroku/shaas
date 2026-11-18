# Image publication

Every push to `master` runs `.github/workflows/docker.yml`, which publishes the shaas image. It must be pullable from **ECR** (the source registry, `s/heroku/shaas/shaas`) and, until it is retired, the **Runtime Registry (RR)**, with identical content.

```
hack/publish-ecr.sh        build each platform ONCE, push <sha>-<arch> to ECR, assemble the index <sha> in ECR
hack/publish-rr.sh         crane copy by digest ECR -> RR (RR-only)
hack/verify-publication.sh fails closed unless the checks below pass
```

This repo uploads no manifests or release metadata. If that is ever added, it must run after `verify-publication.sh`.

Platforms come from the `ARCHITECTURES` secret (comma-separated, e.g. `amd64,arm64`). `publish-ecr.sh` writes `tmp/publication/record.txt` (platform, tag, digest per line); the other two scripts work from that record.

## Paths

| Registry | Repository |
|---|---|
| ECR | `<ecr registry>/s/heroku/shaas/shaas` (from `terraform/data/container-images.csv` in `heroku/runtime-ecr-terraform`) |
| RR | `$RUNTIME_REGISTRY/s/$SERVICE_USERNAME/shaas` (unchanged) |

The two paths differ; the mapping is explicit in the workflow, not inferred.

## What `verify-publication` checks

For ECR and RR: every recorded tag (`<sha>-<arch>` and `<sha>`) resolves to the recorded digest; the index has exactly the recorded platforms with the recorded child digests; every child manifest is readable; RR has the same index and children as ECR. Any read error other than "not found" (403, auth) is a failure. Failures print the image, expected reference, reason, owner and this doc.

## One build

Each platform image is built once, by `docker buildx build --push` straight to ECR, and the index is made from those pushed images by digest (`docker buildx imagetools create`). RR is copied from ECR with `crane copy`; RR content is never copied to ECR and nothing is rebuilt to publish elsewhere. Previously the workflow built, pushed and assembled the index in RR only.

## Metadata tags

Images are built with `--provenance=false --sbom=false`, matching the previous `docker build` output: no attestation, SBOM or signature manifests exist, so the index holds only platform children. If attestations are added later, `publish-rr.sh` must copy them too.

## Retries

Both repos are immutable. Re-running the workflow for the same commit never rebuilds a platform image that is already in ECR and reuses an index that already exists with the expected children (a conflicting index is a failure). `publish-rr.sh` accepts a tag already present with the same digest and fails on a tag with different content. A partially published release therefore cannot be advertised, and can be completed by re-running.

## Tests

`hack/test/publication_test.sh` runs the scripts against a fake registry (stub `crane` and `docker`): complete publication, retries, partial publication, conflicting tags, missing platforms, unreadable manifests and 403s. It runs with shellcheck on pull requests (`.github/workflows/publication-tests.yml`).

## Retiring RR

1. Delete the RR login and "Publish to Runtime Registry (RR)" steps and the `RR_REPO` env in `docker.yml`.
2. Delete `hack/publish-rr.sh`.
3. `verify-publication.sh` keeps working with no RR repo argument.
