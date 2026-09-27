#!/usr/bin/env bash
# =============================================================
#  Ali Telematika ISP — Aplikasi Manajemen ISP
#  INSTALLER SATU FILE (self-extracting)
#
#  Berisi: binary backend (linux) + frontend siap pakai +
#          service systemd + reverse proxy nginx + SSL otomatis.
#  Server TIDAK perlu Go / Node.js — cukup bash, dan installer
#  akan memasang nginx & certbot sendiri.
#
#  CARA PAKAI (unggah file ini ke VPS lalu):
#    chmod +x install-alitelematika.sh
#    sudo ./install-alitelematika.sh
#
#  Installer akan bertanya:
#    1. Domain aplikasi (mis. isp.alitelmatara.id — kosong = IP saja)
#    2. Email untuk sertifikat SSL
#    3. Aktifkan SSL? (y/n)
#
#  MODE NON-INTERAKTIF:
#    sudo APP_DOMAIN=isp.contoh.id SSL_EMAIL=a@b.id ./install-alitelematika.sh
#    (tambahkan SSL=0 untuk tanpa SSL)
#
#  Opsi khusus: --info (lihat versi build), --extract-only DIR
#  (ekstrak isi paket tanpa menginstal — untuk pemeriksaan).
# =============================================================
set -euo pipefail

APP_NAME="Ali Telematika ISP — Manajemen ISP"
APP_VERSION="__APP_VERSION__"
APP_BUILD="__APP_BUILD__"
APP_ARCH="__APP_ARCH__"
PAYLOAD_SHA256="__PAYLOAD_SHA256__"
APP_DIR=/opt/isp-app
NGINX_SITE=/etc/nginx/sites-available/isp-app.conf

BOLD=$'\033[1m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; CYAN=$'\033[36m'; RESET=$'\033[0m'
say()  { printf "%s\n" "${CYAN}==>${RESET} $*"; }
ok()   { printf "%s\n" "${GREEN} ✓${RESET} $*"; }
fail() { printf "%s\n" "${YELLOW}ERROR:${RESET} $*" >&2; exit 1; }
ask()  {
  local prompt="$1" def="${2:-}" re
  printf '%s' "${YELLOW}?${RESET} ${BOLD}${prompt}${RESET}"
  [ -n "$def" ] && printf ' %s[%s]%s' "$RESET" "$def" "$YELLOW"
  printf ': '
  read -r re
  printf '%s' "${re:-$def}"
}

# ---------- opsi khusus ----------
if [ "${1:-}" = "--info" ]; then
  echo "Aplikasi : $APP_NAME"
  echo "Versi    : $APP_VERSION"
  echo "Build    : $APP_BUILD"
  echo "Arsitektur binary: linux/$APP_ARCH"
  echo "Checksum : $PAYLOAD_SHA256"
  exit 0
fi

EXTRACT_DIR=""
if [ "${1:-}" = "--extract-only" ]; then
  EXTRACT_DIR="${2:?--extract-only butuh direktori tujuan}"
  mkdir -p "$EXTRACT_DIR"
fi

# ---------- 0. prasyarat & jawaban ----------
if [ -z "$EXTRACT_DIR" ]; then
  [ "$(id -u)" -eq 0 ] || fail "jalankan sebagai root: sudo $0"
  [ "$(uname -s)" = "Linux" ] || fail "installer ini untuk Linux (uname: $(uname -s))"
  [ "$(uname -m)" = "$APP_ARCH" ] || fail "arsitektur server $(uname -m) tidak cocok dengan binary ($APP_ARCH)"
fi
command -v bash   >/dev/null || fail "bash tidak ditemukan"
command -v base64 >/dev/null || fail "base64 tidak ditemukan (coreutils)"
command -v tar    >/dev/null || fail "tar tidak ditemukan"
command -v sha256sum >/dev/null || fail "sha256sum tidak ditemukan (coreutils)"

echo ""
printf '%s\n' "${BOLD}═══════════════════════════════════════════════════════════${RESET}"
printf '%s\n' "${BOLD}  $APP_NAME${RESET}"
printf '%s\n' "${BOLD}  Installer satu file · v${APP_VERSION} · build ${APP_BUILD}${RESET}"
printf '%s\n' "${BOLD}═══════════════════════════════════════════════════════════${RESET}"
echo ""

DOMAIN="${APP_DOMAIN:-}"
EMAIL="${SSL_EMAIL:-}"
SSL_CHOICE="${SSL:-}"

if [ -z "$EXTRACT_DIR" ]; then
  if [ -z "$DOMAIN" ]; then
    DOMAIN="$(ask 'Domain aplikasi (mis. isp.alitelmatara.id — kosong = pakai IP saja)' '')"
  fi
  DOMAIN="$(echo "$DOMAIN" | tr '[:upper:]' '[:lower:]' | xargs)"
  if [ -n "$DOMAIN" ]; then
    echo "$DOMAIN" | grep -Eq '^([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$' \
      || fail "format domain tidak valid: $DOMAIN"
    if [ -z "$EMAIL" ]; then
      EMAIL="$(ask "Email untuk sertifikat SSL Let's Encrypt" '')"
    fi
    if [ -z "$SSL_CHOICE" ]; then
      if [ -n "$EMAIL" ]; then SSL_CHOICE=y; else SSL_CHOICE=n; fi
      SSL_CHOICE="$(ask "Aktifkan SSL (HTTPS) untuk $DOMAIN?" "$SSL_CHOICE")"
    fi
    case "$SSL_CHOICE" in y|Y|yes|YES|1) SSL=1 ;; *) SSL=0 ;; esac
    [ "$SSL" -eq 1 ] && [ -z "$EMAIL" ] && fail "SSL butuh email; isi SSL_EMAIL atau jawab pertanyaan email"
  else
    SSL=0
  fi
  echo ""
  say "Ringkasan jawaban:"
  if [ -n "$DOMAIN" ]; then ok "Domain : $DOMAIN"; else ok "Domain : (tidak diisi — akses via IP, tanpa SSL)"; fi
  if [ "$SSL" -eq 1 ]; then ok "SSL    : AKTIF (Let's Encrypt, email $EMAIL)"; else ok "SSL    : tidak aktif"; fi
  echo ""
else
  SSL=0; DOMAIN=""
fi

# ---------- 1. ekstrak & verifikasi paket ----------
say "[1/5] Mengekstrak paket aplikasi…"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
SCRIPT_PATH="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"
awk 'f{print} /^#__PAYLOAD_BELOW__$/{f=1}' "$SCRIPT_PATH" | base64 -d > "$TMP/payload.tar.gz"
[ -s "$TMP/payload.tar.gz" ] || fail "paket kosong / file rusak"

GOT_SHA="$(sha256sum "$TMP/payload.tar.gz" | awk '{print $1}')"
if [ "$GOT_SHA" != "$PAYLOAD_SHA256" ]; then
  fail "checksum tidak cocok! File mungkin rusak/tidak lengkap saat diunggah. Unduh ulang installer."
fi
ok "checksum payload cocok ($GOT_SHA)"

TARGET_DIR="${EXTRACT_DIR:-$APP_DIR}"
mkdir -p "$TARGET_DIR"
tar xzf "$TMP/payload.tar.gz" -C "$TARGET_DIR"
[ -f "$TARGET_DIR/isp-api" ] && [ -f "$TARGET_DIR/web/index.html" ] \
  || fail "isi paket tidak lengkap (butuh isp-api dan web/index.html)"
chmod +x "$TARGET_DIR/isp-api"
ok "terekstrak ke $TARGET_DIR (binary + $(find "$TARGET_DIR/web" -type f | wc -l) file web)"

if [ -n "$EXTRACT_DIR" ]; then
  ok "mode --extract-only: berhenti setelah ekstrak (tidak menginstal apa pun)."
  exit 0
fi

# ---------- 2. dependensi runtime ----------
say "[2/5] Menyiapkan dependensi runtime (nginx/certbot)…"
export DEBIAN_FRONTEND=noninteractive
if command -v apt-get >/dev/null 2>&1; then
  apt-get update -y
  apt-get install -y curl nginx
  command -v python3 >/dev/null 2>&1 || apt-get install -y python3
  if [ "$SSL" -eq 1 ]; then apt-get install -y certbot; fi
else
  command -v nginx  >/dev/null || fail "nginx tidak ditemukan — pasang manual lalu jalankan ulang"
  [ "$SSL" -eq 1 ] && { command -v certbot >/dev/null || fail "certbot tidak ditemukan — pasang manual lalu jalankan ulang"; }
fi

# ---------- 3. service systemd ----------
say "[3/5] Memasang service systemd (isp-api)…"
mkdir -p "$APP_DIR/data"
id -u www-data >/dev/null 2>&1 || useradd -r -s /usr/sbin/nologin www-data
chown -R www-data:www-data "$APP_DIR"
cat > /etc/systemd/system/isp-api.service <<'UNIT'
[Unit]
Description=Aplikasi Manajemen ISP - Ali Telematika (Golang API + Frontend)
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/isp-app
ExecStart=/opt/isp-app/isp-api -addr 127.0.0.1:8080 -db /opt/isp-app/data/isp.db -static /opt/isp-app/web
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload
systemctl enable isp-api >/dev/null 2>&1
systemctl restart isp-api
sleep 2
systemctl is-active --quiet isp-api || { journalctl -u isp-api --no-pager -n 20; fail "service isp-api tidak berjalan"; }
HTTP_CODE="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/ || true)"
[ "$HTTP_CODE" = "200" ] || fail "aplikasi tidak merespons di 127.0.0.1:8080 (kode: $HTTP_CODE)"
ok "service isp-api aktif dan merespons (127.0.0.1:8080)"

# ---------- 4. nginx ----------
say "[4/5] Memasang reverse proxy nginx…"
mkdir -p /var/www/letsencrypt
if [ -n "$DOMAIN" ]; then SERVER_NAME="$DOMAIN www.$DOMAIN"; else SERVER_NAME="_"; fi

cat > "$NGINX_SITE" <<NGINX
# Dibuat otomatis oleh installer satu file — file ini ditimpa setiap instalasi.
server {
    listen 80;
    listen [::]:80;
    server_name ${SERVER_NAME};
    server_tokens off;

    client_max_body_size 10m;

    location ^~ /.well-known/acme-challenge/ {
        root /var/www/letsencrypt;
        default_type "text/plain";
    }

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }

    location = /favicon.ico { access_log off; log_not_found off; }
}
NGINX

ln -sf "$NGINX_SITE" /etc/nginx/sites-enabled/isp-app.conf
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx
if [ -n "$DOMAIN" ]; then ok "nginx aktif (HTTP — $DOMAIN)"; else ok "nginx aktif (HTTP via IP)"; fi

# ---------- 5. SSL ----------
FINAL_URL="http://$(hostname -I 2>/dev/null | awk '{print $1}')/"
if [ "$SSL" -eq 1 ]; then
  say "[5/5] Menerbitkan sertifikat SSL Let's Encrypt untuk $DOMAIN…"
  DOMAIN_IP="$(getent hosts "$DOMAIN" | awk '{print $1}' | head -1 || true)"
  SERVER_IP="$(hostname -I | awk '{print $1}')"
  if [ -n "$DOMAIN_IP" ] && [ "$DOMAIN_IP" != "$SERVER_IP" ]; then
    fail "DNS $DOMAIN mengarah ke $DOMAIN_IP, bukan IP server ini ($SERVER_IP). Perbaiki DNS lalu jalankan ulang installer."
  fi
  certbot certonly --webroot -w /var/www/letsencrypt \
    -d "$DOMAIN" -d "www.$DOMAIN" \
    -m "$EMAIL" --agree-tos -n --keep-until-expiring \
    || certbot certonly --webroot -w /var/www/letsencrypt -d "$DOMAIN" \
         -m "$EMAIL" --agree-tos -n --keep-until-expiring \
    || fail "gagal menerbitkan sertifikat. Cek: DNS sudah benar? Port 80 terbuka di provider?"

  cat > "$NGINX_SITE" <<NGINX
# Dibuat otomatis oleh installer satu file — file ini ditimpa setiap instalasi.
# HTTP -> HTTPS (validasi ACME untuk pembaruan sertifikat tetap lewat sini)
server {
    listen 80;
    listen [::]:80;
    server_name ${SERVER_NAME};
    server_tokens off;

    location ^~ /.well-known/acme-challenge/ {
        root /var/www/letsencrypt;
        default_type "text/plain";
    }
    location / { return 301 https://\$host\$request_uri; }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${SERVER_NAME};
    server_tokens off;

    ssl_certificate     /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;

    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header Referrer-Policy "same-origin" always;

    client_max_body_size 10m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_read_timeout 60s;
    }

    location = /favicon.ico { access_log off; log_not_found off; }
}
NGINX
  nginx -t
  systemctl reload nginx
  systemctl enable --now certbot.timer 2>/dev/null || true
  FINAL_URL="https://$DOMAIN/"
  ok "SSL aktif: $FINAL_URL (perpanjangan otomatis via certbot.timer)"
else
  say "[5/5] SSL dilewati (tidak dipilih / tanpa domain)."
fi

# ---------- firewall & selesai ----------
if command -v ufw >/dev/null 2>&1 && ufw status 2>/dev/null | grep -q "Status: active"; then
  ufw allow 'Nginx Full' >/dev/null 2>&1 || true
  ok "ufw: Nginx Full diizinkan"
fi

echo ""
printf '%s\n' "${GREEN}${BOLD}═══════════════════════════════════════════════════════════${RESET}"
printf '%s\n' "${GREEN}${BOLD} SELESAI! $APP_NAME${RESET}"
printf '%s\n' "${GREEN}${BOLD}═══════════════════════════════════════════════════════════${RESET}"
echo ""
echo "   Aplikasi  : ${BOLD}${FINAL_URL}${RESET}"
[ "$SSL" -eq 1 ] && echo "   SSL       : aktif (Let's Encrypt, auto-renew)"
echo "   Login     : admin / admin123   ${YELLOW}(WAJIB diganti segera di Pengaturan → Keamanan)${RESET}"
echo ""
echo "   Perintah penting:"
echo "     systemctl status isp-api      — status aplikasi"
echo "     journalctl -u isp-api -f      — log aplikasi"
echo "     certbot renew --dry-run       — coba pembaruan sertifikat"
echo ""
echo "   Data & backup : $APP_DIR/data/isp.db"
echo "   Konfig nginx  : $NGINX_SITE (dibuat otomatis)"
echo ""
if [ "$SSL" -eq 0 ]; then
  echo "   ${YELLOW}Catatan: SSL nonaktif. Aktifkan kapan saja dengan menjalankan ulang"
  echo "   installer dan isi domain:${RESET} sudo APP_DOMAIN=domainmu.id SSL_EMAIL=emailmu ./$0"
  echo ""
fi
exit 0
#__PAYLOAD_BELOW__
