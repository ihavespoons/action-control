#!/bin/sh
set -e

# Extract the command (enforce, report, export) and arguments
CMD="$1"
REPO="$3"
OUTPUT_FORMAT="$5"
GITHUB_TOKEN="$7"
POLICY_CONTENT="$9"

# Export GitHub token as environment variable
if [ -n "$GITHUB_TOKEN" ]; then
  export ACTION_CONTROL_GITHUB_TOKEN="$GITHUB_TOKEN"
else
  echo "GitHub token not provided"
  exit 1
fi

# Check if policy content is provided
if [ -n "$POLICY_CONTENT" ]; then
  echo "Using policy content from command line argument"
  # Create a temporary file for the policy content
  TEMP_POLICY_FILE="/tmp/policy.yaml"
  echo "$POLICY_CONTENT" > "$TEMP_POLICY_FILE"
  
  # Execute with temporary file and ignore local policy flag
  exec /app/action-control "$CMD" --repo "$REPO" --output "$OUTPUT_FORMAT" --policy "$TEMP_POLICY_FILE" --ignore-local-policy
else
  echo "No policy content provided. Please provide policy content."
  exit 1
fi