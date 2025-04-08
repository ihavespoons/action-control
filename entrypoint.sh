#!/bin/sh

# Handle command and non-conditional args
CMD="action-control $1 --org $2 --repo $4 --output $6 --policy $8"

# Run the command
eval $CMD