# Gemini text generation

This is a text generation tool that uses Gemini AI to generate text based on prompts files.
Utilizes different locations to speed up the process.

## Setup

Adds promts files in the prompts INPUT_DIR.
It must have .prompt extension. After generation the output will be saved in the OUTPUT_DIR
with the same name but without the .prompt extension.

Google API key you can get here: https://console.cloud.google.com/apis/credentials

## Usage

./gemini_turbo --project=1234567890 --cred=./cred/proj-12346-7464f599b1c7.json

## Options help

./gemini_turbo --help
