#!/usr/bin/env bash
set -euo pipefail

PORT="${1:-${DRAGONFLY_PORT:-6379}}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
TEMP_DIR="${ROOT_DIR}/temp"
ARCHIVE_NAME="dragonfly-x86_64.tar.gz"
DOWNLOAD_URL="${DRAGONFLY_DOWNLOAD_URL:-https://github.com/dragonflydb/dragonfly/releases/download/v1.37.0/${ARCHIVE_NAME}}"
ARCHIVE_PATH="${TEMP_DIR}/${ARCHIVE_NAME}"
BIN_PATH="${TEMP_DIR}/dragonfly"

mkdir -p "${TEMP_DIR}"

download_file() {
  local url="$1"
  local out="$2"

  if command -v curl >/dev/null 2>&1; then
    curl -fL --retry 3 --retry-delay 2 -o "${out}" "${url}"
  elif command -v wget >/dev/null 2>&1; then
    wget -O "${out}" "${url}"
  else
    echo "Neither curl nor wget is installed. Install one of them to download Dragonfly." >&2
    exit 1
  fi
}

if [ ! -x "${BIN_PATH}" ]; then
  echo "Dragonfly executable not found at ${BIN_PATH}"
  echo "Downloading Dragonfly from ${DOWNLOAD_URL} ..."
  download_file "${DOWNLOAD_URL}" "${ARCHIVE_PATH}"
  tar -xzf "${ARCHIVE_PATH}" -C "${TEMP_DIR}"
  if [ -f "${TEMP_DIR}/dragonfly-x86_64" ]; then
    mv "${TEMP_DIR}/dragonfly-x86_64" "${BIN_PATH}"
  fi
  chmod +x "${BIN_PATH}"
fi

echo "Starting Dragonfly on port ${PORT}"
echo "Dragonfly port: ${PORT}"

exec "${BIN_PATH}" --logtostderr --port="${PORT}"
