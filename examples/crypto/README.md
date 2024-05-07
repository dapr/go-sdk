# Dapr Crypto Example with go-sdk

## Steps

### Prepare

- Dapr installed

> In order to run this sample, make sure that OpenSSL is available on your system.

### Running

1. This sample requires a private RSA key and a 256-bit symmetric (AES) key. We will generate them using OpenSSL:

<!-- STEP
name: Generate crypto
expected_stderr_lines:
output_match_mode: substring
background: false
sleep: 5
timeout_seconds: 30
-->

```bash
mkdir -p keys
# Generate a private RSA key, 4096-bit keys
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:4096 -out keys/rsa-private-key.pem
# Generate a 256-bit key for AES
openssl rand -out keys/symmetric-key-256 32
```

<!-- END_STEP -->

2. Run the Go service app with Dapr:

<!-- STEP
name: Run crypto example
expected_stdout_lines:
  - '== APP == Encrypted the message, got 856 bytes'
  - '== APP == Decrypted the message, got 24 bytes'
  - '== APP == The secret is "passw0rd"'
  - '== APP == Wrote encrypted data to encrypted.out'
  - '== APP == Wrote decrypted data to decrypted.out.jpg'
  - "Exited App successfully"
expected_stderr_lines:
output_match_mode: substring
sleep: 30
timeout_seconds: 90
-->

```bash
dapr run --app-id crypto --resources-path ./components/ -- go run .
```

<!-- END_STEP -->

### Cleanup

`ctrl + c` to stop execution

```bash
dapr stop --app-id crypto
(lsof -i:8080 | grep crypto) | awk '{print $2}' | xargs  kill
```

## Result

```shell
== APP == Encrypted the message, got 856 bytes
== APP == Decrypted the message, got 24 bytes
== APP == The secret is "passw0rd"
== APP == Wrote encrypted data to encrypted.out
== APP == Wrote decrypted data to decrypted.out.jpg
```
