#!/bin/bash
# Helper script to start Barracuda with GSC credentials from .env file

# Load .env file if it exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Run barracuda serve with any additional arguments
exec barracuda serve "$@"

