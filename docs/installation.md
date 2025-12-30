# Installation

fyaml can be installed using several methods. Choose the one that best fits your environment.

## Quick Install (Linux/macOS)

The fastest way to install fyaml on Linux or macOS:

```bash
curl -sSL https://raw.githubusercontent.com/jksmth/fyaml/main/install.sh | bash
```

This script downloads the latest release for your platform and installs it to `/usr/local/bin/fyaml`.

## From Source (Go)

If you have Go installed, you can install fyaml directly:

```bash
go install github.com/jksmth/fyaml@latest
```

This installs the latest version to `$GOPATH/bin` or `$HOME/go/bin` (depending on your Go version).

### Requirements

- Go 1.25 or later

## From Pre-built Binaries

Download the latest release from the [GitHub releases page](https://github.com/jksmth/fyaml/releases).

### Linux/macOS

```bash
# Download and extract
curl -L https://github.com/jksmth/fyaml/releases/latest/download/fyaml_linux_amd64.tar.gz | tar xz

# Make executable
chmod +x fyaml

# Move to PATH (optional)
sudo mv fyaml /usr/local/bin/

# Verify installation
fyaml version
```

### Windows

1. Download the `.zip` file from the [releases page](https://github.com/jksmth/fyaml/releases)
2. Extract the archive
3. Add the extracted directory to your PATH
4. Verify installation: `fyaml version`

## Docker

### Run Directly

You can run fyaml using Docker without installing it locally:

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/jksmth/fyaml:latest pack /workspace/config
```

### Use in Multi-stage Builds

Include fyaml in your Dockerfile for build-time configuration compilation:

```dockerfile
# Build stage - copy fyaml binary
FROM ghcr.io/jksmth/fyaml:latest AS fyaml

# Your application stage
FROM your-base-image:latest
COPY --from=fyaml /fyaml /usr/local/bin/fyaml

# Use fyaml in your build process
COPY config/ /config/
RUN fyaml pack /config > /app/config.yml
```

### Available Tags

- `latest` - Latest stable release
- `v1.0.0` - Specific version (replace with desired version)

## Verify Installation

After installation, verify that fyaml is working:

```bash
fyaml version
```

You should see output like:

```
1.0.3 (commit: c56e30ab7375f56ea0a57944b1354b970e66d7b2, date: 2025-12-29T23:31:56Z)
```

## Verification

All fyaml releases are cryptographically signed with [cosign](https://github.com/sigstore/cosign) using keyless signing. To verify a binary or Docker image:

```bash
# Verify binary checksums
cosign verify-blob \
  --certificate-identity-regexp '^https://github.com/jksmth/fyaml' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt \
  --bundle checksums.txt.sigstore

# Verify Docker image
cosign verify \
  --certificate-identity-regexp '^https://github.com/jksmth/fyaml' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  ghcr.io/jksmth/fyaml:v1.0.0
```

## Troubleshooting

### Command Not Found

If you get a "command not found" error:

1. **Check installation location**: Verify where fyaml was installed
   ```bash
   which fyaml  # Linux/macOS
   where fyaml  # Windows
   ```

2. **Add to PATH**: Ensure the installation directory is in your PATH
   ```bash
   # Linux/macOS - add to ~/.bashrc or ~/.zshrc
   export PATH="$PATH:/usr/local/bin"

   # Or add the Go bin directory
   export PATH="$PATH:$HOME/go/bin"
   ```

3. **Restart terminal**: Close and reopen your terminal after modifying PATH

### Permission Denied

If you get a permission error:

```bash
# Make executable
chmod +x fyaml

# Or install with sudo
sudo mv fyaml /usr/local/bin/
```

### Docker Issues

If Docker commands fail:

1. Ensure Docker is running: `docker ps`
2. Check image availability: `docker pull ghcr.io/jksmth/fyaml:latest`
3. Verify volume mount syntax for your OS (Windows uses different path format)

