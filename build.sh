#!/bin/sh

# Build 
GOOS=linux GOARCH=arm64 go build -o ./releases/linux_arm64/gemini_turbo
GOOS=linux GOARCH=amd64 go build -o ./releases/linux_amd64/gemini_turbo
GOOS=darwin GOARCH=arm64 go build -o ./releases/macos_arm64/gemini_turbo