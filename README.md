# Gemini text generation

This is a text generation tool that uses Gemini AI to generate text based on prompts files.
Utilizes different locations to speed up the process.

## Setup

Adds promts files in the prompts INPUT_DIR.
It must have .prompt extension. After generation the output will be saved in the OUTPUT_DIR
with the same name but without the .prompt extension.

Google API key you can get here: https://console.cloud.google.com/apis/credentials

## Usage:

gemini_turbo [OPTIONS]

Application Options:
--model= Model (default: gemini-1.5-pro-preview-0409) [$GEMINI_MODEL]
--project= Google Project ID [$GOOGLE_PROJECT_ID]
--cred= Google Credential File [$GOOGLE_CRED_FILE]
--in= Input directory (default: ./prompts) [$INPUT_DIR]
--out= Output directory (default: ./out) [$OUTPUT_DIR]
--max_tokens= Max tokens (default: 8000) [$MAX_TOKENS]

Help Options:
-h, --help Show this help message

## Example

./gemini_turbo --project=1234567890 --cred=proj-12346-7464f599b1c7.json
