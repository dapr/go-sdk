# Dapr Binary Store Example with go-sdk

This example uploads, downloads, overwrites and deletes a file with the binary store API.
It uses the in-memory binary store component, so it needs no external storage; the files are
lost when the sidecar stops.

> The binary store API is in alpha and may change.

## Steps

### Prepare

- Dapr installed (v1.19 or higher)

### Running

Run the Go app with Dapr:

<!-- STEP
name: Run binary store example
expected_stdout_lines:
  - 'Uploaded reports/summary.bin (1048576 bytes)'
  - 'Upload without overwrite failed with AlreadyExists'
  - 'Downloaded 1048576 bytes, SHA-256 matches the upload'
  - 'Overwrote the file, new content: "replaced by the Go SDK"'
  - 'Deleted the file, getting it now fails with NotFound'
  - "Exited App successfully"
expected_stderr_lines:
output_match_mode: substring
timeout_seconds: 120
-->

```bash
dapr run --app-id binary-store --resources-path ./components/ -- go run .
```

<!-- END_STEP -->

## Result

```shell
Uploaded reports/summary.bin (1048576 bytes)
Upload without overwrite failed with AlreadyExists
Downloaded 1048576 bytes, SHA-256 matches the upload
Overwrote the file, new content: "replaced by the Go SDK"
Deleted the file, getting it now fails with NotFound
```

## API

- `SetBinaryFileAlpha1(ctx, storeName, fileName, in, overwrite)` reads `in` until `io.EOF` and
  uploads it in chunks. If `overwrite` is false and the file exists, the error has code
  `codes.AlreadyExists`. If reading `in` fails, the upload is canceled.
- `GetBinaryFileAlpha1(ctx, storeName, fileName)` returns an `io.ReadCloser` that streams the file.
  Errors such as `codes.NotFound` are returned by the call itself. Always `Close` the reader;
  closing it early cancels the download.
- `DeleteBinaryFileAlpha1(ctx, storeName, fileName)` deletes the file.
