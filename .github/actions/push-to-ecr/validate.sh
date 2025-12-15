#!/bin/bash
set -euo pipefail

ecr_repo="$1"
github_repo="$2"
github_repo_name=$(echo "$github_repo" | cut -d'/' -f2)

if ! [[ "$ecr_repo" =~ ^s/ ]]; then
  echo "❌ ECR repository must start with 's/': $ecr_repo"
  exit 1
fi

if ! [[ "$ecr_repo" =~ $github_repo_name ]]; then
  echo "❌ ECR repository must contain GitHub repo name '$github_repo_name': $ecr_repo"
  exit 1
fi

echo "✅ Valid ECR repository: $ecr_repo"
