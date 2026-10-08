package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type session struct {
	verifier string
	expires  time.Time
}
type app struct {
	id, secret, origin, owner, tokenURL, userURL string
	client                                       *http.Client
	mu                                           sync.Mutex
	states                                       map[string]session
}

func randomString() string {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func (a *app) headers(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'nonce-blog-oauth'; frame-ancestors 'none'")
}
func (a *app) auth(w http.ResponseWriter, r *http.Request) {
	a.headers(w)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if provider := r.URL.Query().Get("provider"); provider != "" && provider != "github" {
		http.Error(w, "Unsupported provider", 400)
		return
	}
	state, verifier := randomString(), randomString()
	a.mu.Lock()
	for key, value := range a.states {
		if time.Now().After(value.expires) {
			delete(a.states, key)
		}
	}
	if len(a.states) >= 1000 {
		a.mu.Unlock()
		http.Error(w, "Try again later", 429)
		return
	}
	a.states[state] = session{verifier, time.Now().Add(10 * time.Minute)}
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "blog_oauth_state", Value: state, Path: "/cms-oauth/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: 600})
	challenge := sha256.Sum256([]byte(verifier))
	q := url.Values{"client_id": {a.id}, "redirect_uri": {a.origin + "/cms-oauth/callback"}, "scope": {"public_repo"}, "state": {state}, "code_challenge": {base64.RawURLEncoding.EncodeToString(challenge[:])}, "code_challenge_method": {"S256"}}
	http.Redirect(w, r, "https://github.com/login/oauth/authorize?"+q.Encode(), http.StatusFound)
}

func (a *app) exchange(code, verifier string) (string, error) {
	data := url.Values{"client_id": {a.id}, "client_secret": {a.secret}, "code": {code}, "redirect_uri": {a.origin + "/cms-oauth/callback"}, "code_verifier": {verifier}}
	req, _ := http.NewRequest("POST", a.tokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	response, e := a.client.Do(req)
	if e != nil {
		return "", e
	}
	defer response.Body.Close()
	var result struct {
		Token string `json:"access_token"`
	}
	if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result) != nil || result.Token == "" {
		return "", errors.New("token exchange failed")
	}
	req, _ = http.NewRequest("GET", a.userURL, nil)
	req.Header.Set("Authorization", "Bearer "+result.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "oneyearago-blog-editor")
	response, e = a.client.Do(req)
	if e != nil {
		return "", e
	}
	defer response.Body.Close()
	var user struct {
		Login string `json:"login"`
	}
	if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&user) != nil || !strings.EqualFold(user.Login, a.owner) {
		return "", errors.New("account not allowed")
	}
	return result.Token, nil
}

var callbackPage = template.Must(template.New("callback").Parse(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>博客登录完成</title></head><body><p>登录成功，正在返回文章编辑器。</p><script nonce="blog-oauth">
const origin={{.Origin}}, message={{.Message}};
if (window.opener) {
  window.addEventListener('message', function receive(event) {
    if (event.origin !== origin || event.source !== window.opener || event.data !== 'authorizing:github') return;
    window.removeEventListener('message', receive);
    window.opener.postMessage(message, origin); window.close();
  });
  window.opener.postMessage('authorizing:github', origin);
}
</script></body></html>`))

func (a *app) callback(w http.ResponseWriter, r *http.Request) {
	a.headers(w)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	state := r.URL.Query().Get("state")
	cookie, e := r.Cookie("blog_oauth_state")
	if e != nil || state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(cookie.Value)) != 1 {
		http.Error(w, "登录校验失败，请重新登录。", 403)
		return
	}
	a.mu.Lock()
	s, ok := a.states[state]
	delete(a.states, state)
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "blog_oauth_state", Path: "/cms-oauth/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	if !ok || time.Now().After(s.expires) || r.URL.Query().Get("code") == "" {
		http.Error(w, "登录已失效，请重新登录。", 403)
		return
	}
	token, e := a.exchange(r.URL.Query().Get("code"), s.verifier)
	if e != nil {
		http.Error(w, "登录失败，或此账号没有博客编辑权限。", 403)
		return
	}
	payload, _ := json.Marshal(map[string]string{"token": token, "provider": "github"})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if e := callbackPage.Execute(w, map[string]string{"Origin": a.origin, "Message": "authorization:github:success:" + string(payload)}); e != nil {
		log.Print("Could not render login response")
	}
}

func main() {
	origin := strings.TrimRight(os.Getenv("SITE_ORIGIN"), "/")
	u, e := url.Parse(origin)
	if e != nil || u.Scheme != "https" || u.Host == "" || u.Path != "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		log.Fatal("SITE_ORIGIN must be an HTTPS origin")
	}
	a := &app{id: os.Getenv("GITHUB_CLIENT_ID"), secret: os.Getenv("GITHUB_CLIENT_SECRET"), origin: origin, owner: os.Getenv("GITHUB_OWNER"), tokenURL: "https://github.com/login/oauth/access_token", userURL: "https://api.github.com/user", client: &http.Client{Timeout: 15 * time.Second}, states: make(map[string]session)}
	if a.id == "" || a.secret == "" || a.owner == "" {
		log.Fatal("Missing OAuth configuration")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/cms-oauth/auth", a.auth)
	mux.HandleFunc("/cms-oauth/callback", a.callback)
	mux.HandleFunc("/cms-oauth/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	server := &http.Server{Addr: "127.0.0.1:8787", Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	log.Fatal(server.ListenAndServe())
}
