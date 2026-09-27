package main

import (
	"database/sql"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func bearerToken(r *http.Request) string {
	ah := r.Header.Get("Authorization")
	if strings.HasPrefix(ah, "Bearer ") {
		return strings.TrimPrefix(ah, "Bearer ")
	}
	return ""
}

// checkToken memverifikasi tanda tangan, kedaluwarsa, keberadaan user, dan
// memastikan token diterbitkan TIDAK sebelum password terakhir diganti
// (ganti password mematikan semua sesi lama).
func checkToken(db *sql.DB, token string) bool {
	if token == "" {
		return false
	}
	tok, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(GetSecret(db)), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || tok == nil || !tok.Valid {
		return false
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	uid, ok := claims["uid"].(float64)
	if !ok || uid <= 0 {
		return false
	}
	var changed int64
	if err := db.QueryRow(`SELECT pass_changed FROM users WHERE id=?`, int64(uid)).Scan(&changed); err != nil {
		return false
	}
	iat, _ := claims["iat"].(float64)
	return int64(iat) >= changed
}

func issueToken(db *sql.DB, uid int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{"uid": uid, "iat": now.Unix(), "exp": now.Add(7 * 24 * time.Hour).Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(GetSecret(db)))
}

// ===== Pembatasan percobaan login (anti brute force) =====
// 5 kegagalan dalam 5 menit untuk kombinasi IP+username → terkunci 60 detik.

type loginAttempts struct {
	fails   int
	window  time.Time
	blocked time.Time
}

var (
	loginMu    sync.Mutex
	loginState = map[string]*loginAttempts{}
)

func loginKey(r *http.Request, username string) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return host + "|" + strings.ToLower(strings.TrimSpace(username))
}

func loginGate(key string, now time.Time) (bool, int) {
	loginMu.Lock()
	defer loginMu.Unlock()
	st := loginState[key]
	if st == nil || now.After(st.blocked) {
		return true, 0
	}
	return false, int(st.blocked.Sub(now).Seconds()) + 1
}

func loginFailed(key string, now time.Time) {
	loginMu.Lock()
	defer loginMu.Unlock()
	st := loginState[key]
	if st == nil || now.Sub(st.window) > 5*time.Minute {
		st = &loginAttempts{window: now}
		loginState[key] = st
	}
	st.fails++
	if st.fails >= 5 {
		st.blocked = now.Add(60 * time.Second)
		st.fails = 0
		st.window = now
	}
}

func loginSucceeded(key string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	delete(loginState, key)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &in); err != nil {
		httpErr(w, 400, "Permintaan tidak valid")
		return
	}
	key := loginKey(r, in.Username)
	if ok, wait := loginGate(key, time.Now()); !ok {
		w.Header().Set("Retry-After", strconv.Itoa(wait))
		httpErr(w, 429, "Terlalu banyak percobaan gagal. Coba lagi dalam "+strconv.Itoa(wait)+" detik.")
		return
	}
	var (
		id   int64
		hash string
		nama string
	)
	err := h.DB.QueryRow(`SELECT id, pass_hash, nama FROM users WHERE username=?`, in.Username).Scan(&id, &hash, &nama)
	if err == sql.ErrNoRows || (err == nil && bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil) {
		loginFailed(key, time.Now())
		httpErr(w, 401, "Username atau password salah")
		return
	}
	if err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	token, err := issueToken(h.DB, id)
	if err != nil {
		httpErr(w, 500, "Gagal membuat token")
		return
	}
	loginSucceeded(key)
	writeJSON(w, 200, map[string]any{"token": token, "nama": nama, "username": in.Username})
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Lama string `json:"lama"`
		Baru string `json:"baru"`
	}
	if err := readJSON(r, &in); err != nil || in.Baru == "" {
		httpErr(w, 400, "Permintaan tidak valid")
		return
	}
	if len(in.Baru) < 6 {
		httpErr(w, 400, "Password baru minimal 6 karakter")
		return
	}
	var hash string
	err := h.DB.QueryRow(`SELECT pass_hash FROM users WHERE username='admin'`).Scan(&hash)
	if err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Lama)) != nil {
		httpErr(w, 400, "Password lama salah")
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(in.Baru), bcrypt.DefaultCost)
	if err != nil {
		httpErr(w, 500, "Kesalahan server")
		return
	}
	if _, err = h.DB.Exec(`UPDATE users SET pass_hash=?, pass_changed=? WHERE username='admin'`, string(newHash), time.Now().Unix()); err != nil {
		httpErr(w, 500, "Gagal menyimpan password")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// AuthMiddleware memproteksi endpoint API. File frontend (non /api) dibiarkan
// lewat karena otentikasi SPA ditangani di sisi klien; endpoint HTML dokumen
// memverifikasi token sendiri lewat query string (iframe tidak bisa mengirim
// header Authorization).
func AuthMiddleware(next http.Handler, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if !strings.HasPrefix(p, "/api/") || p == "/api/login" {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(p, "/api/dokumen/") && strings.HasSuffix(p, "/html") {
			next.ServeHTTP(w, r)
			return
		}
		if !checkToken(db, bearerToken(r)) {
			httpErr(w, 401, "Tidak terautentikasi")
			return
		}
		next.ServeHTTP(w, r)
	})
}
