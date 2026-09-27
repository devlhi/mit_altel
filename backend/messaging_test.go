package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPhoneAndGatewayValidation(t *testing.T) {
	for raw, want := range map[string]string{"0812-3456-7890": "6281234567890", "+62 812 3456 7890": "6281234567890", "": ""} {
		v, e := normalizePhone(raw)
		if e != nil || v != want {
			t.Errorf("%s: %s %v", raw, v, e)
		}
	}
	for _, raw := range []string{"abc123456789", "62812345678@g.us", "12", "0000"} {
		if _, e := normalizePhone(raw); e == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	for _, u := range []string{"https://example.com/api", "http://127.0.0.1:5000", "http://192.168.1.2:3000"} {
		if _, e := validGatewayURL(u, false); e != nil {
			t.Fatal(e)
		}
	}
	for _, u := range []string{"http://example.com", "https://user:key@example.com", "https://example.com?key=foo", "file:///tmp"} {
		if _, e := validGatewayURL(u, false); e == nil {
			t.Errorf("accepted %s", u)
		}
	}
	for _, u := range []string{"http://localhost:8080", "https://127.0.0.1:8080", "https://example.com/billing"} {
		if _, e := validGatewayURL(u, true); e == nil {
			t.Errorf("accepted public %s", u)
		}
	}
}
func TestMPWAIdempotencyEncryptionAndInvoiceShare(t *testing.T) {
	db, mux := testOperations(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/send-message" || r.Method != "POST" {
			t.Errorf("wrong path/method")
		}
		_ = r.ParseForm()
		if r.Form.Get("api_key") != "private-test-key" || r.Form.Get("sender") != "628111111111" || r.Form.Get("number") != "628222222222" {
			t.Errorf("wrong gateway fields")
		}
		writeJSON(w, 200, map[string]any{"status": true, "data": map[string]any{"key": map[string]string{"id": "MSG-1"}}})
	}))
	defer server.Close()
	c := defaultWAConfig()
	c.BaseURL = server.URL
	c.Enabled = true
	c.Sender = "08111111111"
	c.PublicURL = "https://billing.example.com"
	raw := map[string]any{}
	_ = json.Unmarshal([]byte(jsonText(c)), &raw)
	raw["api_key"] = "private-test-key"
	w := opRequest(t, mux, "PUT", "/api/whatsapp/settings", raw, 200)
	if strings.Contains(w.Body.String(), "private-test-key") {
		t.Fatal("API key exposed")
	}
	var stored string
	_ = db.QueryRow(`SELECT secret FROM wa_config`).Scan(&stored)
	if stored == "" || strings.Contains(stored, "private-test-key") {
		t.Fatal("key not encrypted")
	}
	raw["api_key"] = ""
	opRequest(t, mux, "PUT", "/api/whatsapp/settings", raw, 200)
	var id int64
	_ = db.QueryRow(`SELECT id FROM tagihan LIMIT 1`).Scan(&id)
	c, _, _ = loadWA(db)
	source, e := waSource(db, c, "invoice", id)
	if e != nil {
		t.Fatal(e)
	}
	in := WASendInput{Key: "test-request-key-00001", Kind: "invoice", SourceID: id, Revision: source.Revision, Recipient: "08222222222", Message: source.Message, IncludeLink: true}
	w = opRequest(t, mux, "POST", "/api/whatsapp/send", in, 200)
	var out WAOutbox
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if out.Status != "accepted" || out.ProviderID != "MSG-1" || !out.HasShare {
		t.Fatalf("bad result %s", w.Body.String())
	}
	opRequest(t, mux, "POST", "/api/whatsapp/send", in, 200)
	if calls.Load() != 1 {
		t.Fatal("duplicate message sent")
	}
	changed := in
	changed.Message = "Different"
	opRequest(t, mux, "POST", "/api/whatsapp/send", changed, 409)
	prefix := c.PublicURL + "/share/"
	pos := strings.Index(out.Message, prefix)
	if pos < 0 {
		t.Fatal("missing document link")
	}
	token := out.Message[pos+len(prefix) : pos+len(prefix)+64]
	public := httptest.NewRecorder()
	AuthMiddleware(mux, db).ServeHTTP(public, httptest.NewRequest("GET", "/share/"+token, nil))
	if public.Code != 200 || !strings.Contains(public.Body.String(), source.Reference) || !strings.Contains(public.Body.String(), "PLG-") || public.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("share not served safely")
	}
	opRequest(t, mux, "POST", "/api/whatsapp/history/"+strconvID(out.ID)+"/revoke", nil, 200)
	opRequest(t, mux, "GET", "/share/"+token, nil, 404)
	private := httptest.NewRecorder()
	AuthMiddleware(mux, db).ServeHTTP(private, httptest.NewRequest("GET", "/api/whatsapp/settings", nil))
	if private.Code != 401 {
		t.Fatal("settings are public")
	}
	// An invoice changed after preview cannot be sent with a stale body.
	execTest(t, db, `UPDATE tagihan SET jumlah=jumlah+1 WHERE id=?`, id)
	in.Key = "test-request-key-00002"
	opRequest(t, mux, "POST", "/api/whatsapp/send", in, 409)
	if calls.Load() != 1 {
		t.Fatal("stale invoice was sent")
	}
}
func strconvID(n int64) string { return jsonText(n) }

func TestMPWAResponsesRedirectAndNoRetry(t *testing.T) {
	for _, tc := range []struct {
		name, body, status string
		code               int
	}{{"accepted", `{"status":true}`, "accepted", 200}, {"rejected", `{"status":false,"msg":"SECRET-DO-NOT-LEAK"}`, "failed", 200}, {"bad-key", `SECRET-DO-NOT-LEAK`, "failed", 401}, {"server-error", `{}`, "unknown", 503}, {"malformed", `html`, "unknown", 200}, {"no-status", `{}`, "unknown", 200}} {
		t.Run(tc.name, func(t *testing.T) {
			var count int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				w.WriteHeader(tc.code)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()
			status, _, err := sendMPWA(WAConfig{BaseURL: server.URL, Sender: "628111111111"}, "SECRET-DO-NOT-LEAK", "628222222222", "test", "")
			if status != tc.status || count != 1 || strings.Contains(err, "SECRET-DO-NOT-LEAK") {
				t.Fatalf("result %s %s requests %d", status, err, count)
			}
		})
	}
	var forwarded atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded.Add(1) }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	status, _, _ := sendMPWA(WAConfig{BaseURL: server.URL}, "key", "628222222222", "test", "")
	if status != "failed" || forwarded.Load() != 0 {
		t.Fatal("credential redirect followed")
	}
}
func TestMPWAPDFAndExpiredLinks(t *testing.T) {
	db, mux := testOperations(t)
	c := defaultWAConfig()
	c.PublicURL = "https://billing.example.com"
	execTest(t, db, `INSERT INTO wa_config(id,data,secret) VALUES(1,?,'')`, jsonText(c))
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	part, _ := mw.CreateFormFile("file", "invoice.pdf")
	_, _ = part.Write([]byte("%PDF-1.4\nmock-pdf-for-transport-test"))
	_ = mw.Close()
	req := httptest.NewRequest("POST", "/api/whatsapp/attachment", &b)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var result map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	token := result["attachment_id"]
	w = opRequest(t, mux, "GET", "/share/"+token, nil, 200)
	if w.Header().Get("Content-Type") != "application/pdf" {
		t.Fatal("not PDF")
	}
	execTest(t, db, `UPDATE wa_shares SET expires_at=?`, time.Now().Add(-time.Hour).Unix())
	opRequest(t, mux, "GET", "/share/"+token, nil, 404)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.URL.Path != "/send-media" || r.Form.Get("media_type") != "document" || r.Form.Get("url") != "https://billing.example.com/share/token" || r.Form.Get("caption") != "caption" {
			t.Error("wrong media payload")
		}
		writeJSON(w, 200, map[string]bool{"status": true})
	}))
	defer server.Close()
	status, _, e := sendMPWA(WAConfig{BaseURL: server.URL}, "key", "628222222222", "caption", "https://billing.example.com/share/token")
	if status != "accepted" {
		t.Fatal(e)
	}
}

func TestMPWAInFlightDuplicateAndTimeout(t *testing.T) {
	db, mux := testOperations(t)
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-release
		writeJSON(w, 200, map[string]bool{"status": true})
	}))
	defer gateway.Close()
	c := defaultWAConfig()
	c.Enabled = true
	c.BaseURL = gateway.URL
	c.Sender = "628111111111"
	encrypted, e := sealWA(db, "test-key", false)
	if e != nil {
		t.Fatal(e)
	}
	execTest(t, db, `INSERT INTO wa_config(id,data,secret) VALUES(1,?,?)`, jsonText(c), encrypted)
	in := WASendInput{Key: "concurrent-request-0001", Kind: "custom", Recipient: "628222222222", Message: "test only"}
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/api/whatsapp/send", strings.NewReader(jsonText(in))))
		done <- w
	}()
	<-started
	// Neither a duplicate send nor reads need to wait for the provider.
	w := opRequest(t, mux, "POST", "/api/whatsapp/send", in, 200)
	if !strings.Contains(w.Body.String(), `"status":"sending"`) {
		t.Fatal(w.Body.String())
	}
	opRequest(t, mux, "GET", "/api/whatsapp/settings", nil, 200)
	close(release)
	w = <-done
	if w.Code != 200 || calls.Load() != 1 {
		t.Fatalf("duplicate or failure: %d %d", w.Code, calls.Load())
	}
	old := mpwaTimeout
	mpwaTimeout = 15 * time.Millisecond
	defer func() { mpwaTimeout = old }()
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Millisecond)
		writeJSON(w, 200, map[string]bool{"status": true})
	}))
	defer slow.Close()
	status, _, _ := sendMPWA(WAConfig{BaseURL: slow.URL}, "test-key", "628222222222", "test only", "")
	if status != "unknown" {
		t.Fatalf("timeout must remain uncertain: %s", status)
	}
}
