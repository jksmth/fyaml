#!/usr/bin/env bash

# Install the fyaml CLI tool.
# https://github.com/jksmth/fyaml
#
# Dependencies: curl, cut, sha256sum (Linux) or shasum (macOS)
#
# The version to install and the binary location can be passed in via VERSION and DESTDIR respectively.
#

set -o errexit

function error {
  echo "An error occurred installing the tool."
  echo "The contents of the directory $SCRATCH have been left in place to help to debug the issue."
}

# Verify checksum of downloaded file
function verify_checksum {
	local file="$1"
	local expected_checksum="$2"

	if [ -z "$expected_checksum" ]; then
		echo "WARNING: No checksum found in checksums.txt for this file. Skipping verification." >&2
		return 0
	fi

	echo "Verifying checksum..."
	local actual_checksum

	# Detect and use the appropriate checksum command
	if command -v sha256sum >/dev/null 2>&1; then
		actual_checksum=$(sha256sum "$file" | awk '{print $1}')
	elif command -v shasum >/dev/null 2>&1; then
		actual_checksum=$(shasum -a 256 "$file" | awk '{print $1}')
	else
		echo "ERROR: Neither sha256sum nor shasum found. Cannot verify checksum." >&2
		echo "Please install one of them or skip verification." >&2
		return 1
	fi

	if [ "$expected_checksum" != "$actual_checksum" ]; then
		echo "ERROR: Checksum verification failed!" >&2
		echo "Expected: $expected_checksum" >&2
		echo "Actual:   $actual_checksum" >&2
		echo "The downloaded file may be corrupted or tampered with." >&2
		return 1
	fi

	echo "Checksum verified successfully."
	return 0
}

# Use a function to ensure connection errors don't partially execute when being piped
function install_cli {

	echo "Starting installation."

	# GitHub's URL for the latest release, will redirect.
	GITHUB_BASE_URL="https://github.com/jksmth/fyaml"
	LATEST_URL="${GITHUB_BASE_URL}/releases/latest/"
	DESTDIR="${DESTDIR:-/usr/local/bin}"

	if [ -z "$VERSION" ]; then
		VERSION=$(curl -sLI -o /dev/null -w '%{url_effective}' "$LATEST_URL" | cut -d "v" -f 2)
	fi

	echo "Installing fyaml v${VERSION}"

	# Run the script in a temporary directory that we know is empty.
	SCRATCH=$(mktemp -d || mktemp -d -t 'tmp')
	cd "$SCRATCH"

	trap error ERR

	# Determine release filename.
	case "$(uname)" in
		Linux)
			OS='linux'
		;;
		Darwin)
			OS='darwin'
		;;
		*)
			echo "This operating system is not supported."
			exit 1
		;;
	esac

	case "$(uname -m)" in
		aarch64 | arm64)
			ARCH='arm64'
		;;
		x86_64)
			ARCH="amd64"
		;;
		*)
			echo "This architecture is not supported."
			exit 1
		;;
	esac

	RELEASE_FILENAME="fyaml_${VERSION}_${OS}_${ARCH}.tar.gz"
	RELEASE_URL="${GITHUB_BASE_URL}/releases/download/v${VERSION}/${RELEASE_FILENAME}"
	CHECKSUMS_URL="${GITHUB_BASE_URL}/releases/download/v${VERSION}/checksums.txt"

	# Download checksums.txt
	echo "Downloading checksums..."
	if ! curl --ssl-reqd -sL --retry 3 --fail "${CHECKSUMS_URL}" -o checksums.txt; then
		echo "WARNING: Failed to download checksums.txt. Installation will continue without verification." >&2
		CHECKSUMS_AVAILABLE=false
	else
		CHECKSUMS_AVAILABLE=true
	fi

	# Download the release tarball to a file (not pipe) for verification
	echo "Downloading ${RELEASE_FILENAME}..."
	if ! curl --ssl-reqd -sL --retry 3 --fail "${RELEASE_URL}" -o "${RELEASE_FILENAME}"; then
		echo "ERROR: Failed to download release file: ${RELEASE_FILENAME}" >&2
		echo "URL: ${RELEASE_URL}" >&2
		exit 1
	fi

	# Verify checksum if available
	if [ "$CHECKSUMS_AVAILABLE" = "true" ]; then
		EXPECTED_CHECKSUM=$(grep "${RELEASE_FILENAME}$" checksums.txt | awk '{print $1}')
		if ! verify_checksum "${RELEASE_FILENAME}" "$EXPECTED_CHECKSUM"; then
			echo "ERROR: Checksum verification failed. Aborting installation." >&2
			exit 1
		fi
	fi

	# Extract the tarball
	echo "Extracting..."
	tar zxf "${RELEASE_FILENAME}"

	if [ ! -f "fyaml" ]; then
		echo "ERROR: Binary 'fyaml' not found in archive" >&2
		exit 1
	fi

	echo "Installing to $DESTDIR"
	install fyaml "$DESTDIR"

	command -v fyaml

	# Delete the working directory when the install was successful.
	rm -r "$SCRATCH"
}

install_cli
