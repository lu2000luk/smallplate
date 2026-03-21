#!/usr/bin/env bash
set -euo pipefail

PORT="${1:-${VALKEY_PORT:-6378}}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
TEMP_DIR="${ROOT_DIR}/temp"
VERSION="${VALKEY_VERSION:-8.0.0}"
ARCH="${VALKEY_ARCH:-x86_64}"

mkdir -p "${TEMP_DIR}"

download_file() {
  local url="$1"
  local out="$2"

  if command -v curl >/dev/null 2>&1; then
    curl -fL --retry 3 --retry-delay 2 -o "${out}" "${url}"
  elif command -v wget >/dev/null 2>&1; then
    wget -O "${out}" "${url}"
  else
    echo "Neither curl nor wget is installed. Install one of them to download Valkey." >&2
    exit 1
  fi
}

detect_distro() {
  if [ -n "${VALKEY_DISTRO:-}" ]; then
    printf '%s\n' "${VALKEY_DISTRO}"
    return
  fi

  local codename=""
  if [ -r /etc/os-release ]; then
    codename="$(
      . /etc/os-release
      printf '%s' "${VERSION_CODENAME:-${UBUNTU_CODENAME:-}}"
    )"
  fi

  case "${codename}" in
    noble|jammy|focal|bionic)
      printf '%s\n' "${codename}"
      ;;
    *)
      printf '%s\n' "noble"
      ;;
  esac
}

DISTRO="$(detect_distro)"
ARCHIVE_BASENAME="valkey-${VERSION}-${DISTRO}-${ARCH}"
ARCHIVE_NAME="${ARCHIVE_BASENAME}.tar.gz"
DOWNLOAD_URL="${VALKEY_DOWNLOAD_URL:-https://download.valkey.io/releases/${ARCHIVE_NAME}}"
ARCHIVE_PATH="${TEMP_DIR}/${ARCHIVE_NAME}"
EXTRACTED_DIR="${TEMP_DIR}/${ARCHIVE_BASENAME}"
BIN_PATH="${EXTRACTED_DIR}/bin/valkey-server"

if [ ! -x "${BIN_PATH}" ]; then
  echo "Valkey executable not found at ${BIN_PATH}"
  echo "Downloading Valkey ${VERSION} for ${DISTRO}/${ARCH} from ${DOWNLOAD_URL} ..."
  rm -rf "${EXTRACTED_DIR}"
  download_file "${DOWNLOAD_URL}" "${ARCHIVE_PATH}"
  tar -xzf "${ARCHIVE_PATH}" -C "${TEMP_DIR}"

  if [ ! -x "${BIN_PATH}" ]; then
    echo "Valkey executable still not found after extraction: ${BIN_PATH}" >&2
    exit 1
  fi

  chmod +x "${BIN_PATH}"
fi

echo "Starting Valkey on port ${PORT}"
echo "Valkey build: ${DISTRO}/${ARCH}"
echo "Valkey port: ${PORT}"

exec "${BIN_PATH}" --port "${PORT}"
