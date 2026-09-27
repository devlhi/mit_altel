#!/usr/bin/env bash
# =============================================================
#  Builder Installer Satu File — Ali Telematika ISP
#
#  Dijalankan di komputer DEVELOPMENT (butuh Go >= 1.26 dan
#  Node >= 18). Menghasilkan:
#
#    install-alitelematika.sh   (di root proyek)
#
#  Satu file berisi: binary backend linux + frontend statis +
#  seluruh logika instalasi (systemd, nginx, SSL, prompt domain).
#  Upload file itu saja ke VPS, lalu:
#    chmod +x install-alitelematika.sh && sudo ./install-alitelematika.sh
#
#  Cara pakai:
#    bash deploy/build-installer.sh              (linux/amd64)
#    ARCH=arm64 bash deploy/build-installer.sh   (linux/arm64)
# =============================================================
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARCH="${ARCH:-amd64}"
VERSION="$(cat "$ROOT/deploy/version.txt" 2>/dev/null || echo 1.0.0)"
BUILD="$(date +%Y%m%d-%H%M)"
# amd64 memakai nama utama (kompatibilitas), selain itu diberi akhiran -<arch>
if [ "$ARCH" = "amd64" ]; then
  OUT="$ROOT/install-alitelematika.sh"
else
  OUT="$ROOT/install-alitelematika-${ARCH}.sh"
fi

BOLD=$'\033[1m'; GREEN=$'\033[32m'; CYAN=$'\033[36m'; RESET=$'\033[0m'
say() { printf "%s\n" "${CYAN}==>${RESET} $*"; }
ok()  { printf "%s\n" "${GREEN} ✓${RESET} $*"; }

case "$ARCH" in amd64|arm64) ;; *) echo "ARCH harus amd64 atau arm64"; exit 1 ;; esac

command -v go   >/dev/null || { echo "Go tidak ditemukan"; exit 1; }
command -v node >/dev/null || { echo "Node tidak ditemukan"; exit 1; }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
PKG="$TMP/pkg"
mkdir -p "$PKG"

# 1. Build frontend segar
say "[1/4] Build frontend (Nuxt 3)…"
cd "$ROOT/frontend"
npm ci --no-audit --no-fund 2>/dev/null || npm install --no-audit --no-fund
npm run generate >/dev/null 2>&1 || npm run generate
cp -r .output/public "$PKG/web"

# 2. Cross-compile backend untuk Linux
say "[2/4] Cross-compile backend untuk linux/$ARCH…"
cd "$ROOT/backend"
export GOTOOLCHAIN=auto
CGO_ENABLED=0 GOOS=linux GOARCH="$ARCH" go build -trimpath -ldflags "-s -w" -o "$PKG/isp-api" .

# 3. Paket + checksum
say "[3/4] Mengemas payload (tar.gz + base64 + sha256)…"
tar czf "$TMP/payload.tar.gz" -C "$PKG" isp-api web
SHA="$(sha256sum "$TMP/payload.tar.gz" | awk '{print $1}')"

# 4. Rakit installer satu file
say "[4/4] Merakit install-alitelematika.sh…"
sed -e "s|__APP_VERSION__|$VERSION|g" \
    -e "s|__APP_BUILD__|$BUILD|g" \
    -e "s|__APP_ARCH__|$ARCH|g" \
    -e "s|__PAYLOAD_SHA256__|$SHA|g" \
    "$ROOT/deploy/installer-template.sh" > "$OUT"
base64 -w 0 "$TMP/payload.tar.gz" >> "$OUT"
printf '\n' >> "$OUT"
chmod +x "$OUT"

SIZE_H="$(du -h "$OUT" | awk '{print $1}')"
OUT_SHA="$(sha256sum "$OUT" | awk '{print $1}')"
echo ""
ok "installer siap: ${BOLD}install-alitelematika.sh${RESET} ($SIZE_H, linux/$ARCH, v$VERSION build $BUILD)"
echo "   checksum installer : $OUT_SHA"
echo ""
echo "   Cara pakai di VPS:"
echo "     1) unggah file install-alitelematika.sh ke VPS (scp/git)"
echo "     2) chmod +x install-alitelematika.sh"
echo "     3) sudo ./install-alitelematika.sh   — lalu isi domain & email SSL saat ditanya"
echo ""
