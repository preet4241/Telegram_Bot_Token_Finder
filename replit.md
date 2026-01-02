# Token Hunter Bot

A Telegram bot token generator and tester written in Go.

## Overview

This application generates random Telegram bot tokens and tests them against the Telegram API. When an active token is found, it sends the details to a configured Telegram chat.

## Project Structure

- `main.go` - Main application code with web server and token hunting workers
- `go.mod` - Go module dependencies

## Environment Variables Required

- `BOT_TOKEN` - Your Telegram bot token for sending notifications
- `CHAT_ID` - Telegram chat ID where results will be sent

## Running the Application

The application runs a web server on port 5000 and spawns background workers to generate and test tokens.

## Architecture

- **Web Server**: Simple HTTP server serving a status page at `/` and health check at `/health`
- **Workers**: 50 concurrent workers generating and testing random tokens
- **Notification**: Results sent via Telegram Bot API

## Recent Changes

- 2026-01-02: Initial Replit setup, fixed Go string literals, configured for port 5000
