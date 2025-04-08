#!/bin/sh
# filepath: /Users/ben.gittins/Code/personal/action-control/entrypoint.sh
set -e

# If policy content is provided as an environment variable, write it to a temporary file
if [ -n "$ACTION_CONTROL_POLICY_CONTENT" ]; then
  echo "Using policy content from environment variable"
  # We'll set the --ignore-local-policy flag to true
  IGNORE_FLAG="--ignore-local-policy"
else
  echo "No policy content provided in environment variable. Please provide ACTION_CONTROL_POLICY_CONTENT."
  echo "Exiting..."
  exit 1
fi

# Export the GitHub token with the expected name
if [ -n "$GITHUB_TOKEN" ]; then
  export ACTION_CONTROL_GITHUB_TOKEN="$GITHUB_TOKEN"
fi

# Extract the command (enforce, report, export) from the first argument
CMD="$1"
shift

# Run the action-control command with appropriate arguments - using fixed positions
# We know the command structure is enforce --repo REPO --output FORMAT
exec /app/action-control "$CMD" --repo "$2" --output "$4" $IGNORE_FLAG