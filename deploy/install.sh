#!/usr/bin/env bash
# =============================================================
#  Installer Aplikasi Manajemen ISP — Ali Telematika (Ubuntu/Debian)
#
#  CARA PAKAI (interaktif — installer akan bertanya):
#    sudo bash deploy/install.sh
#
#  Pertanyaan yang muncul:
#    1. Domain aplikasi      (mis. isp.alitelmatara.id — kosongkan jika
#                             hanya pakai IP, SSL tidak bisa untuk IP)
#    2. Email untuk SSL      (Let's Encrypt butuh email untuk notifikasi)
#    3. Aktifkan SSL?        (y/n — otomatis "y" bila domain & email terisi)
#
#  MODE NON-INTERAKTIF (untuk script/CI):
#    sudo APP_DOMAIN=isp.contoh.id SSL_EMAIL=admin@contoh.id \
#         bash deploy/install.sh
#    Set SSL=0 untuk memaksa tanpa SSL meski domain terisi.
#
#  Yang dilakukan installer:
#   1. Pasang dependensi build (Go 1.26.6, Node 20, nginx, certbot)
#   2. Build backend  -> /opt/isp-app/isp-api  (satu binary, statis)
#   3. Build frontend -> /opt/isp-app/web      (file statis)
#   4. Pasang service systemd (isp-api, autostart + auto-restart)
#   5. Pasang reverse proxy nginx
#   6. Jika domain diisi: terbitkan sertifikat SSL Let's Encrypt
#      (gratis) + pengalihan HTTP->HTTPS + HSTS
# =============================================================
set -euo pipefail

APP_DIR=/opt/isp-app
SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_TARBALL_VER="go1.26.6"
NGINX_SITE=/etc/nginx/sites-available/isp-app.conf

# ---------- warna & helper ----------
BOLD=$'\033[1m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; CYAN=$'\033[36m'; RESET=$'\033[0m'
say()  { printf "%s\n" "${CYAN}==>${RESET} $*"; }
ok()   { printf "%s\n" "${GREEN} ✓${RESET} $*"; }
ask()  { # ask "Pertanyaan" "default" -> mencetak jawaban (atau default)
  local prompt="$1" def="${2:-}" re
  printf '%s' "${YELLOW}?${RESET} ${BOLD}${prompt}${RESET}"
  [ -n "$def" ] && printf ' %s[%s]%s' "$RESET" "$def" "$YELLOW"
  printf ': '
  read -r re
  printf '%s' "${re:-$def}"
}
fail() { printf "%s\n" "${YELLOW}ERROR:${RESET} $*" >&2; exit 1; }

[ "$(id -u)" -eq 0 ] || fail "jalankan sebagai root: sudo bash deploy/install.sh"
[ -f "$SRC_DIR/backend/main.go" ] || fail "jalankan dari folder root proyek (yang berisi backend/ dan frontend/)"

# ---------- 0. Kumpulkan jawaban (env bisa menimpa) ----------
echo ""
printf '%s\n' "${BOLD}═══════════════════════════════════════════════════════════${RESET}"
printf '%s\n' "${BOLD}  Instalasi Ali Telematika ISP — Manajemen ISP${RESET}"
printf '%s\n' "${BOLD}═══════════════════════════════════════════════════════════${RESET}"
echo ""

DOMAIN="${APP_DOMAIN:-}"
if [ -z "$DOMAIN" ]; then
  DOMAIN="$(ask 'Domain aplikasi (mis. isp.alitelmatara.id — kosong = pakai IP saja)' '')"
fi
DOMAIN="$(echo "$DOMAIN" | tr '[:upper:]' '[:lower:]' | xargs)"

EMAIL="${SSL_EMAIL:-}"
SSL_CHOICE="${SSL:-}"

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
  if [ -n "$SSL_CHOICE" ]; then
    case "$SSL_CHOICE" in y|Y|yes|YES|1)
      fail "SSL membutuhkan domain (sertifikat tidak bisa diterbitkan untuk IP)"
    ;; esac
  fi
fi

echo ""
say "Ringkasan jawaban:"
if [ -n "$DOMAIN" ]; then ok "Domain : $DOMAIN"; else ok "Domain : (tidak diisi — akses via IP, tanpa SSL)"; fi
if [ "$SSL" -eq 1 ]; then ok "SSL    : AKTIF (Let's Encrypt, email $EMAIL)"; else ok "SSL    : tidak aktif"; fi
echo ""

# ---------- 1. Dependensi ----------
say "[1/7] Menyiapkan dependensi build…"
export DEBIAN_FRONTEND=noninteractive
export GOTOOLCHAIN=auto
apt-get update -y
apt-get install -y curl ca-certificates nginx
command -v python3 >/dev/null 2>&1 || apt-get install -y python3
if [ "$SSL" -eq 1 ]; then apt-get install -y certbot python3-certbot-nginx; fi

install_go() {
  local arch
  case "$(uname -m)" in
    x86_64)  arch=amd64 ;;
    aarch64) arch=arm64 ;;
    *) fail "arsitektur $(uname -m) tidak didukung" ;;
  esac
  say "    Mengunduh Go ${GO_TARBALL_VER} untuk linux/${arch}…"
  curl -fsSL "https://go.dev/dl/${GO_TARBALL_VER}.linux-${arch}.tar.gz" -o /tmp/go.tgz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tgz
  echo 'export PATH=/usr/local/go/bin:$PATH' > /etc/profile.d/go-path.sh
  export PATH=/usr/local/go/bin:$PATH
}

if ! command -v go >/dev/null 2>&1; then
  install_go
else
  GO_VER="$(go version | awk '{print $3}' | sed 's/^go//')"
  GO_MAJOR="${GO_VER%%.*}"; GO_MINOR="$(echo "$GO_VER" | cut -d. -f2)"
  if [ "$GO_MAJOR" -lt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 22 ]; }; then
    say "    Go $GO_VER terlalu lama (butuh >= 1.22), memasang versi baru…"
    install_go
  fi
fi

NODE_MAJOR=0
command -v node >/dev/null 2>&1 && NODE_MAJOR="$(node -v | sed 's/^v\([0-9]*\).*/\1/')"
if [ "$NODE_MAJOR" -lt 18 ]; then
  say "    Memasang Node.js 20.x…"
  curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
  apt-get install -y nodejs
fi
ok "Go: $(go version) · Node: $(node -v)"

# ---------- 2. Build backend ----------
say "[2/7] Build backend (Golang)…"
systemctl stop isp-api 2>/dev/null || true
mkdir -p "$APP_DIR"
cd "$SRC_DIR/backend"
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "$APP_DIR/isp-api" .

# ---------- 3. Build frontend ----------
say "[3/7] Build frontend (Nuxt 3)…"
cd "$SRC_DIR/frontend"
npm ci --no-audit --no-fund 2>/dev/null || npm install --no-audit --no-fund
npm run generate
rm -rf "$APP_DIR/web"
cp -r .output/public "$APP_DIR/web"

# ---------- 4. Data & service ----------
say "[4/7] Menyiapkan direktori data & service systemd…"
mkdir -p "$APP_DIR/data"
id -u www-data >/dev/null 2>&1 || useradd -r -s /usr/sbin/nologin www-data
chown -R www-data:www-data "$APP_DIR"
cp "$SRC_DIR/deploy/isp-api.service" /etc/systemd/system/isp-api.service
systemctl daemon-reload
systemctl enable isp-api >/dev/null 2>&1
systemctl restart isp-api
sleep 2
systemctl is-active --quiet isp-api || { journalctl -u isp-api --no-pager -n 20; fail "service isp-api tidak berjalan"; }
ok "service isp-api aktif (127.0.0.1:8080)"

# ---------- 5. Nginx ----------
say "[5/7] Memasang reverse proxy nginx…"
mkdir -p /var/www/letsencrypt
if [ -n "$DOMAIN" ]; then SERVER_NAME="$DOMAIN www.$DOMAIN"; else SERVER_NAME="_"; fi

cat > "$NGINX_SITE" <<NGINX
# Dibuat otomatis oleh deploy/install.sh — file ini ditimpa setiap instalasi.
server {
    listen 80;
    listen [::]:80;
    server_name ${SERVER_NAME};
    server_tokens off;

    client_max_body_size 10m;

    # Validasi pembaruan sertifikat Let's Encrypt tetap lewat HTTP
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

# ---------- 6. SSL ----------
FINAL_URL="http://$(hostname -I 2>/dev/null | awk '{print $1}')/"
if [ "$SSL" -eq 1 ]; then
  say "[6/7] Menerbitkan sertifikat SSL Let's Encrypt untuk $DOMAIN…"
  DOMAIN_IP="$(getent hosts "$DOMAIN" | awk '{print $1}' | head -1 || true)"
  SERVER_IP="$(hostname -I | awk '{print $1}')"
  if [ -n "$DOMAIN_IP" ] && [ "$DOMAIN_IP" != "$SERVER_IP" ]; then
    fail "DNS $DOMAIN mengarah ke $DOMAIN_IP, bukan IP server ini ($SERVER_IP). Perbaiki DNS lalu jalankan ulang installer."
  fi
  mkdir -p /var/www/letsencrypt
  certbot certonly --webroot -w /var/www/letsencrypt \
    -d "$DOMAIN" -d "www.$DOMAIN" \
    -m "$EMAIL" --agree-tos -n --keep-until-expiring \
    || certbot certonly --webroot -w /var/www/letsencrypt -d "$DOMAIN" \
         -m "$EMAIL" --agree-tos -n --keep-until-expiring \
    || fail "gagal menerbitkan sertifikat. Cek: DNS sudah benar? Port 80 terbuka di provider?"

  cat > "$NGINX_SITE" <<NGINX
# Dibuat otomatis oleh deploy/install.sh — file ini ditimpa setiap instalasi.
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

    # Keamanan tambahan
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
  say "[6/7] SSL dilewati (tidak dipilih / tanpa domain)."
fi

# ---------- 7. Firewall & selesai ----------
say "[7/7] Menyelesaikan…"
if command -v ufw >/dev/null 2>&1 && ufw status 2>/dev/null | grep -q "Status: active"; then
  ufw allow 'Nginx Full' >/dev/null 2>&1 || true
  ok "ufw: Nginx Full diizinkan"
fi

echo ""
printf '%s\n' "${GREEN}${BOLD}═══════════════════════════════════════════════════════════${RESET}"
printf '%s\n' "${GREEN}${BOLD} SELESAI! Aplikasi Manajemen ISP — Ali Telematika${RESET}"
printf '%s\n' "${GREEN}${BOLD}═══════════════════════════════════════════════════════════${RESET}"
echo ""
echo "   Aplikasi  : ${BOLD}${FINAL_URL}${RESET}"
[ "$SSL" -eq 1 ] && echo "   SSL       : aktif (Let's Encrypt, auto-renew)"
echo "   Login     : admin / admin123   ${YELLOW}(WAJIB diganti segera di Pengaturan → Keamanan)${RESET}"
echo ""
echo "   Perintah penting:"
echo "     systemctl status isp-api      — status aplikasi"
echo "     journalctl -u isp-api -f      — log aplikasi"
echo "     systemctl reload nginx        — setelah edit konfigurasi nginx"
echo "     certbot renew --dry-run       — coba pembaruan sertifikat"
echo ""
echo "   Data & backup : $APP_DIR/data/isp.db"
echo "   Konfig nginx  : $NGINX_SITE (dibuat otomatis installer)"
echo ""
if [ "$SSL" -eq 0 ]; then
  echo "   ${YELLOW}Catatan: SSL nonaktif. Aktifkan kapan saja dengan menjalankan ulang"
  echo "   installer dan isi domain:${RESET} sudo APP_DOMAIN=domainmu.id SSL_EMAIL=emailmu bash deploy/install.sh"
  echo ""
fi
