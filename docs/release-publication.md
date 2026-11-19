# Image publication

Every push to `master` runs `.github/workflows/docker.yml`, which publishes the shaas image. It must be pullable from **ECR** (the source registry, `s/heroku/shaas/shaas`) and, until it is retired, the **Runtime Registry (RR)**, with identical content.

The index, copy and verify steps are shared actions in `heroku/kubernetes` (`actions/publication/`; its README says what each checks). The workflow references them at `master`.

```
build + push each platform        docker buildx build, ONCE per platform -> ECR <sha>-<arch>   (repo-local)
publication/create-index          index <sha> in ECR from the platform digests
publication/copy-to-rr            copy the index and the <sha>-<arch> images by digest ECR -> RR (RR-only)
publication/verify                fails closed unless the index and children are correct in ECR and RR
```

This repo uploads no manifests or release metadata. If that is ever added, it must run after `verify`.

Platforms come from the `ARCHITECTURES` secret (comma-separated, e.g. `amd64,arm64`); `verify` checks exactly those platforms. RR keeps its `<sha>-<arch>` and `<sha>` tags. `verify` checks the index (and so each child); the per-arch tags are copied but not separately verified. Images are built with `--provenance=false --sbom=false`, so there are no attestation children in the index.

## Paths

| Registry | Repository |
|---|---|
| ECR | `<ecr registry>/s/heroku/shaas/shaas` (from `terraform/data/container-images.csv` in `heroku/runtime-ecr-terraform`) |
| RR | `$RUNTIME_REGISTRY/s/$SERVICE_USERNAME/shaas` (unchanged) |

The paths differ. Consumers must use the ECR path; the mapping is explicit in the workflow (`source-prefix` and `rr-prefix`), not inferred.

## Retries

ECR tags are immutable. A platform image already in ECR for the commit is reused, not rebuilt; `create-index` and `copy-to-rr` accept identical existing content and fail on a tag with different content.

## Retiring RR

Delete the "Copy images to Runtime Registry (RR)" step and the `rr-prefix` input of the verify step in `docker.yml`.
