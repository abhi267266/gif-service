# File Upload API - Cloudflare R2

A simple Go Fiber API for uploading files to Cloudflare R2 using MinIO client.

## Setup

1. **Create .env file** (copy from .env.example):
   ```bash
   cp .env.example .env
   ```

2. **Install dependencies** (already done):
   ```bash
   go mod download
   ```

## Running the API

```bash
go run .
```

Or build and run the binary:

```bash
go build -o api-server
./api-server
```

The server will start on port 8080 (or the port specified in your .env file).

## API Endpoints

### GET /
Returns API information and available endpoints.

**Response:**
```json
{
  "message": "File Upload API - Cloudflare R2",
  "endpoints": {
    "POST /upload": "Upload a file to R2"
  }
}
```

### POST /upload
Upload a file to Cloudflare R2.

**Request:**
- Method: POST
- Content-Type: multipart/form-data
- Body: form-data with key `file` containing the file to upload

**Example using cURL:**
```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@/path/to/your/file.jpg"
```

**Example using httpie:**
```bash
http -f POST localhost:8080/upload file@/path/to/your/file.jpg
```

**Success Response:**
```json
{
  "message": "File uploaded successfully",
  "filename": "file.jpg",
  "size": 12345,
  "bucket": "video-bucket",
  "etag": "\"abc123...\""
}
```

**Error Responses:**

No file uploaded (400):
```json
{
  "error": "No file uploaded"
}
```

Upload failed (500):
```json
{
  "error": "Failed to upload file to R2"
}
```

## Environment Variables

See `.env.example` for all required configuration:

- `R2_ENDPOINT` - Your R2 endpoint URL
- `R2_ACCESS_KEY` - R2 access key
- `R2_SECRET_KEY` - R2 secret key
- `R2_BUCKET` - R2 bucket name
- `R2_USE_SSL` - Use SSL (true/false)
- `PORT` - API server port (default: 8080)

## Features

- ✅ File upload to Cloudflare R2
- ✅ Support for large files (up to 100MB)
- ✅ CORS enabled
- ✅ Request logging
- ✅ Content type detection
- ✅ Environment variable configuration

## Testing

Test the API with a sample file:

```bash
# Create a test file
echo "Hello R2!" > test.txt

# Upload it
curl -X POST http://localhost:8080/upload \
  -F "file=@test.txt"
```
