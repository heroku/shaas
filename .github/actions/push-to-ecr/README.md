# Push to ECR

Authenticates with Amazon ECR using OIDC and pushes pre-built Docker images.

## Usage

### Single-architecture image

```yaml
- uses: ./.github/actions/push-to-ecr
  with:
    ecr-repository-path: s/heroku/shaas/shaas
    image-tag: ${{ github.sha }}
    source-image-tags: shaas
```

### Multi-architecture image

```yaml
- uses: ./.github/actions/push-to-ecr
  with:
    ecr-repository-path: s/heroku/shaas/shaas
    image-tag: ${{ github.sha }}
    source-image-tags: amd64,arm64
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `ecr-repository-path` | ECR repository path (e.g., `s/heroku/shaas/shaas`) | Yes | - |
| `image-tag` | Image tag (e.g., commit SHA) | Yes | - |
| `source-image-tags` | Comma-separated list of local image tags to push | Yes | - |
| `aws-region` | AWS region | No | `us-east-1` |
| `aws-role-arn` | AWS IAM role ARN for OIDC | No | `arn:aws:iam::970597968373:role/ecr-repo-gha-oidc-iam-role` |

## Outputs

| Output | Description |
|--------|-------------|
| `image-uri` | Full ECR image URI that was pushed |

## Prerequisites

- Docker images must be built and tagged locally before calling this action
- For single-arch: image tagged with the name in `source-image-tags`
- For multi-arch: images tagged with architecture names (e.g., `amd64`, `arm64`)
- GitHub Actions workflow must have `id-token: write` permission for OIDC

## Behavior

**Single-arch:** Tags and pushes one image to ECR.

**Multi-arch:** Tags and pushes architecture-specific images, then creates and pushes a manifest list.
