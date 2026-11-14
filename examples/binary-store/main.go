/*
Copyright 2026 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dapr "github.com/dapr/go-sdk/client"
)

const (
	// Name of the binary store component in ./components
	storeName = "binarystore"
	fileName  = "reports/summary.bin"
)

func main() {
	ctx := context.Background()

	client, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize the Dapr client: %v", err)
	}
	defer client.Close()

	// Upload 1 MiB from a reader; the SDK sends it in chunks
	data := bytes.Repeat([]byte("0123456789abcdef"), 64<<10)
	err = client.SetBinaryFileAlpha1(ctx, storeName, fileName, bytes.NewReader(data), false)
	if err != nil {
		log.Fatalf("Failed to upload the file: %v", err)
	}
	fmt.Printf("Uploaded %s (%d bytes)\n", fileName, len(data))

	// Without overwrite, uploading to an existing file fails
	err = client.SetBinaryFileAlpha1(ctx, storeName, fileName, strings.NewReader("other data"), false)
	if status.Code(err) != codes.AlreadyExists {
		log.Fatalf("Expected an AlreadyExists error, got: %v", err)
	}
	fmt.Println("Upload without overwrite failed with AlreadyExists")

	// Download the file as a stream and compare it with what was uploaded
	downloaded := readFile(ctx, client)
	if sha256.Sum256(downloaded) != sha256.Sum256(data) {
		log.Fatalf("The downloaded file doesn't match the upload")
	}
	fmt.Printf("Downloaded %d bytes, SHA-256 matches the upload\n", len(downloaded))

	// Replace the file
	err = client.SetBinaryFileAlpha1(ctx, storeName, fileName, strings.NewReader("replaced by the Go SDK"), true)
	if err != nil {
		log.Fatalf("Failed to overwrite the file: %v", err)
	}
	fmt.Printf("Overwrote the file, new content: %q\n", readFile(ctx, client))

	// Delete the file; getting it again fails with NotFound
	err = client.DeleteBinaryFileAlpha1(ctx, storeName, fileName)
	if err != nil {
		log.Fatalf("Failed to delete the file: %v", err)
	}
	_, err = client.GetBinaryFileAlpha1(ctx, storeName, fileName)
	if status.Code(err) != codes.NotFound {
		log.Fatalf("Expected a NotFound error, got: %v", err)
	}
	fmt.Println("Deleted the file, getting it now fails with NotFound")
}

func readFile(ctx context.Context, client dapr.Client) []byte {
	r, err := client.GetBinaryFileAlpha1(ctx, storeName, fileName)
	if err != nil {
		log.Fatalf("Failed to get the file: %v", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		log.Fatalf("Failed to read the file: %v", err)
	}
	return data
}
