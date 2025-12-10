# Push Image to ECR

A GitHub Action that authenticates with Amazon ECR, validates image tags, and pushes Docker images to an ECR repository. Supports both single-architecture and multi-architecture builds.

## Features

- ✅ Automatic ECR authentication using AWS OIDC
- ✅ Image tag validation according to ECR requirements
- ✅ Single-architecture image pushes
- ✅ Multi-architecture manifest creation and push
- ✅ Outputs the full image URI for downstream use

## Prerequisites

1. **AWS IAM Role**: An IAM role configured for OIDC authentication with GitHub Actions
   - Default role: `arn:aws:iam::970597968373:role/ecr-repo-gha-oidc-iam-role`
   - The role must have permissions to push to the ECR repository

2. **ECR Repository**: The ECR repository must exist in the specified AWS region

3. **Docker Image**: The Docker image must be built before using this action

4. **GitHub OIDC**: Your workflow must have `permissions` configured for OIDC:
   ```yaml
   permissions:
     id-token: write
     contents: read
   ```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `ecr-repository` | ECR repository name (e.g., `s/heroku/widgetsrus/widgetsrus`). **Must start with `s/` and contain the GitHub repository name.** | Yes | - |
| `image-tag` | Image tag to use (e.g., commit SHA or version number) | Yes | - |
| `aws-region` | AWS region for ECR | No | `us-east-1` |
| `aws-role-arn` | AWS IAM role ARN for OIDC authentication | No | `arn:aws:iam::970597968373:role/ecr-repo-gha-oidc-iam-role` |
| `architectures` | Comma-separated list of architectures for multi-arch build (e.g., `"amd64,arm64"`). **Omit this input or leave empty for single-architecture builds** (default behavior). | No | `''` |

## Outputs

| Output | Description |
|--------|-------------|
| `image-uri` | Full ECR image URI that was pushed (e.g., `970597968373.dkr.ecr.us-east-1.amazonaws.com/s/heroku/widgetsrus/widgetsrus:v1.0.0`) |
| `registry` | ECR registry URL (e.g., `970597968373.dkr.ecr.us-east-1.amazonaws.com`) |

## ECR Repository Requirements

The ECR repository name must follow a specific format that is validated by this action:

- **Must start with `s/`**: The repository path must begin with `s/`
- **Must contain GitHub repository name**: The repository path must include the name of the GitHub repository (the part after `owner/` in `owner/repo`)

### Valid Repository Names
- `s/heroku/myapp/myapp` (if GitHub repo is `owner/myapp`)
- `s/team/widgetsrus/widgetsrus` (if GitHub repo is `owner/widgetsrus`)
- `s/org/project-name/project-name` (if GitHub repo is `owner/project-name`)

### Invalid Repository Names
- `myapp/myapp` (doesn't start with `s/`)
- `s/heroku/different-name` (doesn't contain the GitHub repo name)
- `heroku/myapp/myapp` (doesn't start with `s/`)

The action automatically extracts your GitHub repository name from `${{ github.repository }}` and validates that it appears in the ECR repository path.

## Image Tag Requirements

ECR has strict requirements for image tags. This action validates tags before pushing:

- **Length**: 1-128 characters
- **First character**: Must be alphanumeric (a-z, A-Z, 0-9)
- **Allowed characters**: Letters, numbers, hyphens (`-`), underscores (`_`), periods (`.`), forward slashes (`/`)
- **Cannot start with**: Period (`.`) or hyphen (`-`)

### Valid Tags
- `v1.0.0`
- `abc123`
- `feature/my-branch`
- `2024-01-15`
- `sha-abc123def456`

### Invalid Tags
- `.invalid` (starts with period)
- `-invalid` (starts with hyphen)
- `tag with spaces` (contains spaces)
- `tag@special` (contains @ symbol)

## Usage

### Single-Architecture Build (Default)

For a standard single-architecture Docker image, **simply omit the `architectures` input** (or leave it empty):

```yaml
name: Build and Push to ECR

on:
  push:
    branches: [main]

permissions:
  id-token: write
  contents: read

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Build Docker image
        run: |
          docker build -t myapp:${{ github.sha }} .
      
      - name: Push to ECR
        uses: ./.github/actions/push-to-ecr
        with:
          ecr-repository: s/heroku/myapp/myapp
          image-tag: ${{ github.sha }}
          aws-region: us-east-1
      
      - name: Use pushed image URI
        run: |
          echo "Image pushed to: ${{ steps.push-to-ecr.outputs.image-uri }}"
```

**Note**: For single-arch builds, the action expects the image to be tagged as `{repository-name}:{image-tag}`. In the example above, if your repository is `myapp`, the image should be tagged as `myapp:${{ github.sha }}`.

### Multi-Architecture Build

For multi-architecture builds (e.g., amd64 and arm64):

```yaml
name: Build and Push Multi-Arch to ECR

on:
  push:
    branches: [main]

permissions:
  id-token: write
  contents: read

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Build multi-arch images
        run: |
          docker buildx build --platform linux/amd64 -t amd64 .
          docker buildx build --platform linux/arm64 -t arm64 .
      
      - name: Push to ECR
        id: push-to-ecr
        uses: ./.github/actions/push-to-ecr
        with:
          ecr-repository: s/heroku/myapp/myapp
          image-tag: ${{ github.sha }}
          aws-region: us-east-1
          architectures: amd64,arm64
      
      - name: Use pushed image URI
        run: |
          echo "Multi-arch image pushed to: ${{ steps.push-to-ecr.outputs.image-uri }}"
```

**Note**: For multi-arch builds, the action expects images to be tagged with architecture names (e.g., `amd64`, `arm64`). The action will create a manifest list that combines all architectures.

### Using Custom AWS Role

If you need to use a different IAM role:

```yaml
- name: Push to ECR
  uses: ./.github/actions/push-to-ecr
  with:
    ecr-repository: s/heroku/myapp/myapp
    image-tag: ${{ github.sha }}
    aws-region: us-west-2
    aws-role-arn: arn:aws:iam::123456789012:role/my-custom-role
```

## Common Use Cases

### Tag with Version Number

```yaml
- name: Push to ECR
  uses: ./.github/actions/push-to-ecr
  with:
    ecr-repository: s/heroku/myapp/myapp
    image-tag: v1.2.3
```

### Tag with Branch Name

```yaml
- name: Push to ECR
  uses: ./.github/actions/push-to-ecr
  with:
    ecr-repository: s/heroku/myapp/myapp
    image-tag: ${{ github.ref_name }}
```

### Tag with Commit SHA (Short)

```yaml
- name: Push to ECR
  uses: ./.github/actions/push-to-ecr
  with:
    ecr-repository: s/heroku/myapp/myapp
    image-tag: ${{ github.sha }}
```

## Troubleshooting

### Invalid Repository Error

If you see an error about invalid ECR repository format:
1. Ensure the repository path starts with `s/`
2. Verify the repository path contains your GitHub repository name
3. Check that you're using the correct format: `s/{team}/{repo-name}/{repo-name}`

### Invalid Tag Error

If you see an error about invalid image tags, ensure your tag meets ECR requirements:
- Starts with an alphanumeric character
- Contains only allowed characters (a-z, A-Z, 0-9, `-`, `_`, `.`, `/`)
- Is between 1-128 characters

### Authentication Errors

If you encounter authentication errors:
1. Verify the IAM role ARN is correct
2. Ensure the role has the necessary ECR permissions
3. Check that OIDC is properly configured in your workflow (`permissions` section)

### Image Not Found

For single-arch builds, ensure the image is tagged correctly:
- Tag format: `{repository-name}:{image-tag}`
- Example: If repository is `myapp` and tag is `v1.0.0`, image should be `myapp:v1.0.0`

For multi-arch builds, ensure images are tagged with architecture names:
- Tag format: `{architecture}` (e.g., `amd64`, `arm64`)

## See Also

- [AWS ECR Documentation](https://docs.aws.amazon.com/ecr/)
- [GitHub Actions OIDC](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)

