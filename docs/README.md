# Asset Manager Project

## Overview
This project is designed to serve as a robust alternative for serving Nitro client assets.

## Purpose
Traditionally, Nitro client assets might be served directly from a local folder. This service aims to improve upon that architecture by:
1.  **Encouraging the use of S3 Storage Engines**: Moving towards object storage provides better scalability, reliability, and separating concerns.
2.  **Robust Service**: Providing a dedicated Go service to handle asset requests efficiently.

## Features
- Alternative to local file serving.
- Integration with S3-compatible storage solutions.
- **Secure**: Requires an API Key for access.

## Configuration
The application requires configuration via environment variables or a `.env` file.

### Required Settings
- `SERVER_API_KEY`: **CRITICAL**. You must provide this key to interact with the API.

See `.env.example` for all available configuration options.
