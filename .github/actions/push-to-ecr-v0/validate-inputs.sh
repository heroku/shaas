#!/bin/bash
set -euo pipefail

# Validate ECR repository format
validate_ecr_repository() {
  local ecr_repo="$1"
  local github_repo_name="$2"
  
  # Repository must start with s/
  if ! echo "$ecr_repo" | grep -qE '^s/'; then
    echo "❌ Invalid ECR repository: $ecr_repo"
    echo ""
    echo "ECR repository requirements:"
    echo "  - Must start with 's/'"
    echo "  - Must contain the GitHub repository name: $github_repo_name"
    echo "  - Example format: s/heroku/$github_repo_name/$github_repo_name"
    exit 1
  fi
  
  # Repository must contain the GitHub repo name
  if ! echo "$ecr_repo" | grep -qE "$github_repo_name"; then
    echo "❌ Invalid ECR repository: $ecr_repo"
    echo ""
    echo "ECR repository requirements:"
    echo "  - Must start with 's/'"
    echo "  - Must contain the GitHub repository name: $github_repo_name"
    echo "  - Example format: s/heroku/$github_repo_name/$github_repo_name"
    exit 1
  fi
  
  echo "✅ Valid ECR repository: $ecr_repo"
}

# Validate image tag format
validate_image_tag() {
  local tag="$1"
  
  # ECR tag requirements:
  # - 1-128 characters
  # - Must start with alphanumeric character
  # - Can contain: a-z, A-Z, 0-9, hyphens, underscores, periods, forward slashes
  if ! echo "$tag" | grep -qE '^[a-zA-Z0-9][a-zA-Z0-9._/-]{0,127}$'; then
    echo "❌ Invalid image tag: $tag"
    echo ""
    echo "ECR tag requirements:"
    echo "  - Must be 1-128 characters"
    echo "  - Must start with alphanumeric character (a-z, A-Z, 0-9)"
    echo "  - Can contain: letters, numbers, hyphens, underscores, periods, slashes"
    echo "  - Cannot start with period or hyphen"
    exit 1
  fi
  
  echo "✅ Valid image tag: $tag"
}

# Main validation
main() {
  local ecr_repository="$1"
  local image_tag="$2"
  local github_repository="$3"
  
  # Extract GitHub repository name (the part after owner/)
  local github_repo_name
  github_repo_name=$(echo "$github_repository" | cut -d'/' -f2)
  
  validate_ecr_repository "$ecr_repository" "$github_repo_name"
  validate_image_tag "$image_tag"
}

# Run validation
main "$@"

