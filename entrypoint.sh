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

# Extract the command (enforce, report, export) from the first argument
CMD="$1"
shift

# Run the action-control command with appropriate arguments
exec /app/action-control "$CMD" --org "$2" --repo "$4" --output "$6" $IGNORE_FLAG