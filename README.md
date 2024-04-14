# Gemini text generation

This is a text generation tool that uses Gemini AI to generate text based on prompts files.
Utilizes different locations to speed up the process.

## Setup

Adds promts files in the prompts INPUT_DIR.
It must have .prompt extension. After generation the output will be saved in the OUTPUT_DIR
with the same name but without the .prompt extension.

Set options in environment variables.

Example:

```
GOOGLE_PROJECT_ID=1234567890
GOOGLE_CRED_FILE=./cred/proj-12346-7464f599b1c7.json
GEMINI_MODEL=gemini-1.5-pro-preview-0409
MAX_TOKENS=8000
INPUT_DIR=./prompts
OUTPUT_DIR=./out
```

You can set it in start.sh file and run it ./start.sh
