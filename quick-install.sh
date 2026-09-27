#!/usr/bin/env bash
# =============================================================
#  Ali Telematika ISP — QUICK INSTALL (satu perintah)
#
#  Salin-tempel SATU baris ini di VPS Ubuntu/Debian:
#
#    bash <(curl -fsSL https://raw.githubusercontent.com/devlhi/mit_altel/main/quick-install.sh)
#
#  Script ini otomatis:
#    - mendeteksi arsitektur server (amd64 / arm64)
#    - mengunduh installer lengkap (binary + frontend + konfigurasi)
#    - menjalankannya sebagai root dengan prompt interaktif:
#        ? Domain aplikasi      (mis. isp.alitelmatara.id)
#        ? Email untuk SSL
#        ? Aktifkan SSL? (y/n)
#
#  Mode non-interaktif:
#    sudo APP_DOMAIN=isp.contoh.id SSL_EMAIL=a@b.id \
#      bash quick-install.sh        # atau lewati script ini:
#    sudo APP_DOMAIN=isp.contoh.id SSL_EMAIL=a@b.id \
#      bash <(curl -fsSL https://raw.githubusercontent.com/devlhi/mit_altel/main/quick-install.sh)
#
#  Variabel override (opsional): BASE_RAW (mirror lain), INSTALLER_URL (file persis).
# =============================================================
set -euo pipefail

REPO_RAW="${BASE_RAW:-https://raw.githubusercontent.com/devlhi/mit_altel/main}"

BOLD=$'\033[1m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; CYAN=$'\033[36m'; RESET=$'\033[0m'
say()  { printf "%s\n" "${CYAN}==>${RESET} $*" >&2; }
ok()   { printf "%s\n" "${GREEN} ✓${RESET} $*" >&2; }
fail() { printf "%s\n" "${YELLOW}ERROR:${RESET} $*" >&2; exit 1; }

command -v curl >/dev/null 2>&1 || command -v wget >/dev/null 2>&1 \
  || fail "butuh curl atau wget (apt-get install curl)"

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH=amd64; FILE="install-alitelematika.sh" ;;
  aarch64|arm64) ARCH=arm64; FILE="install-alitelematika-arm64.sh" ;;
  *) fail "arsitektur tidak didukung: $ARCH (didukung: amd64, arm64)" ;;
esac
URL="${INSTALLER_URL:-$REPO_RAW/$FILE}"

TMP="$(mktemp /tmp/alitelematika-install.XXXXXX.sh)"
trap 'rm -f "$TMP"' EXIT

say "Mengunduh installer Ali Telematika ISP (linux/$ARCH)…"
say "Sumber: $URL"
if command -v curl >/dev/null 2>&1; then
  curl -fSL --retry 3 "$URL" -o "$TMP"
else
  wget -q --tries=3 -O "$TMP" "$URL"
fi
[ -s "$TMP" ] || fail "unduhan kosong — periksa koneksi/repo"
ok "terunduh: $(du -h "$TMP" | awk '{print $1}') · sha256 $(sha256sum "$TMP" | awk '{print $1}')"

# Teruskan variabel mode non-interaktif melewati sudo (sudo me-reset env)
PASSED=()
[ -n "${APP_DOMAIN:-}" ] && PASSED+=(env "APP_DOMAIN=$APP_DOMAIN")
[ -n "${SSL_EMAIL:-}" ]  && PASSED+=("SSL_EMAIL=$SSL_EMAIL")
[ -n "${SSL:-}" ]        && PASSED+=("SSL=$SSL")

if [ "$(id -u)" -eq 0 ]; then
  exec bash "$TMP" "$@"
else
  command -v sudo >/dev/null 2>&1 || fail "jalankan sebagai root atau pasang sudo"
  say "Meminta akses root (sudo)…"
  exec sudo "${PASSED[@]}" bash "$TMP" "$@"
fi
