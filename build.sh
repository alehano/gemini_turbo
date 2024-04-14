#!/bin/sh

# Build 
GOOS=linux GOARCH=arm64 go build -o ./releases/linux_arm64/gemini_turbo
GOOS=linux GOARCH=amd64 go build -o ./releases/linux_amd64/gemini_turbo
GOOS=darwin GOARCH=arm64 go build -o ./releases/macos_arm64/gemini_turbo

# Copy files
cp README.md ./releases/linux_arm64/
cp start.sh ./releases/linux_arm64/
mkdir  ./releases/linux_arm64/out
mkdir  ./releases/linux_arm64/prompts
cp -r ./prompts/* ./releases/linux_arm64/prompts/

cp README.md ./releases/linux_amd64/
cp start.sh ./releases/linux_amd64/
mkdir  ./releases/linux_amd64/out
mkdir  ./releases/linux_amd64/prompts
cp -r ./prompts/* ./releases/linux_amd64/prompts/

cp README.md ./releases/macos_arm64/
cp start.sh ./releases/macos_arm64/
mkdir  ./releases/macos_arm64/out
mkdir  ./releases/macos_arm64/prompts
cp -r ./prompts/* ./releases/macos_arm64/prompts/