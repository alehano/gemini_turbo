# Gemini Turbo text generation

This is a text generation tool that uses Gemini AI to generate text based on prompts files.
Utilizes different locations and performs jobs in parallel to speed up the process.

## Setup

Adds promts files in the prompts INPUT_DIR.
It must have .prompt extension. After generation the output will be saved in the OUTPUT_DIR
with the same name but without the .prompt extension.

You need Vortex API key to use this tool.
Credentials file you can get here: https://console.cloud.google.com/apis/credentials

## Usage:

```
Usage:
  gemini_turbo [OPTIONS]

Application Options:
      --model=      Model (default: gemini-1.5-pro-preview-0409) [$GEMINI_MODEL]
      --project=    Google Project ID [$GOOGLE_PROJECT_ID]
      --cred=       Google Credential File [$GOOGLE_CRED_FILE]
      --in=         Input directory (default: ./prompts) [$INPUT_DIR]
      --out=        Output directory (default: ./out) [$OUTPUT_DIR]
      --max_tokens= Max tokens (default: 8000) [$MAX_TOKENS]
      --workers=    Workers (default: 500) [$WORKERS]
      --Delay=      Delay between requests in ms. Should be more than 60000 / req per min limit (5 by default) / number of
                    locations (default: 500ms) [$DELAY]
      --Timeout=    Timeout for each request (default: 300s) [$TIMEOUT]
      --Limit=      Limit files to process. Can be used for testing. 0 means no limit (default: 0) [$LIMIT]

Help Options:
  -h, --help        Show this help message

```

## Example with minimum parameters:

```
./gemini_turbo --project=1234567890 --cred=proj-12346-7464f599b1c7.json
```
