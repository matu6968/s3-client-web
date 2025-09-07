# S3 Client Web

This is a WebGUI for [s3-client](https://github.com/matu6968/s3-client)

## Prerequisites

- Go (1.23 or later)

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/matu6968/s3-client-web
   ```

2. Go to the project directory:
   ```
   cd s3-client-web
   ```

3. Build the binary:
   ```
   go build -o s3-client-web
   ```

## Configuration

In the .env file this is the only thing you can set

```
PORT=8080
```

### For this to even work

`s3-client` is now statically linked, you just need to configure `s3-client` as usual as per [the instructions](https://github.com/matu6968/s3-client?tab=readme-ov-file#configuration)
