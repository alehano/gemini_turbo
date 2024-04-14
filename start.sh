#!/bin/sh

set -e

# Set environment variables

export GOOGLE_PROJECT_ID=1234567890
export GOOGLE_CRED_FILE=./cred/proj-12346-7464f599b1c7.json
export GEMINI_MODEL=gemini-1.5-pro-preview-0409
export MAX_TOKENS=8000
export INPUT_DIR=./prompts
export OUTPUT_DIR=./out

# run 

./gemini_turbo