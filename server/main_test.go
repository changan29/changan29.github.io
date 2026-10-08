package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func testApp(login string) (*app, func()) {
	github := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			r.ParseForm()
			if r.Form.Get("code_verifier") != "verifier" {
				http.Error(w, "bad PKCE", 400)
				return
			}
			w.Write([]byte(`{"access_token":"test-token"}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "missing token", 401)
			return
		}
		w.Write([]byte(`{"login":"` + login + `"}`))
	}))
	a := &app{id: "client", secret: "secret", origin: "https://www.oneyearago.me", owner: "changan29", tokenURL: github.URL + "/token", userURL: github.URL + "/user", client: github.Client(), states: map[string]session{"valid-state": {"verifier", time.Now().Add(time.Minute)}}}
	return a, github.Close
}

func callbackRequest(state, cookie string) *http.Request {
	r := httptest.NewRequest("GET", "https://www.oneyearago.me/cms-oauth/callback?code=code&state="+url.QueryEscape(state), nil)
	r.AddCookie(&http.Cookie{Name: "blog_oauth_state", Value: cookie})
	return r
}

func TestCallbackOwnerAndReplay(t *testing.T) {
	a, close := testApp("changan29")
	defer close()
	w := httptest.NewRecorder()
	a.callback(w, callbackRequest("valid-state", "valid-state"))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "authorization:github:success:") {
		t.Fatal("owner could not log in")
	}
	if !strings.Contains(w.Body.String(), "event.origin !== origin") || !strings.Contains(w.Body.String(), "event.source !== window.opener") {
		t.Fatal("opener validation missing")
	}
	w = httptest.NewRecorder()
	a.callback(w, callbackRequest("valid-state", "valid-state"))
	if w.Code != 403 {
		t.Fatal("state replay accepted")
	}
}

func TestCallbackRejectsOtherAccount(t *testing.T) {
	a, close := testApp("someone-else")
	defer close()
	w := httptest.NewRecorder()
	a.callback(w, callbackRequest("valid-state", "valid-state"))
	if w.Code != 403 || strings.Contains(w.Body.String(), "test-token") {
		t.Fatal("unauthorized account received a token")
	}
}

func TestCallbackRejectsBadOrExpiredState(t *testing.T) {
	a, close := testApp("changan29")
	defer close()
	for _, pair := range [][2]string{{"valid-state", "wrong-cookie"}, {"unknown", "unknown"}, {"", ""}} {
		w := httptest.NewRecorder()
		a.callback(w, callbackRequest(pair[0], pair[1]))
		if w.Code != 403 {
			t.Fatal("bad state accepted")
		}
	}
	a.states["valid-state"] = session{"verifier", time.Now().Add(-time.Minute)}
	w := httptest.NewRecorder()
	a.callback(w, callbackRequest("valid-state", "valid-state"))
	if w.Code != 403 {
		t.Fatal("expired state accepted")
	}
}

func TestAuthUsesSecureCookieAndPKCE(t *testing.T) {
	a, close := testApp("changan29")
	defer close()
	w := httptest.NewRecorder()
	a.auth(w, httptest.NewRequest("GET", "https://www.oneyearago.me/cms-oauth/auth?provider=github", nil))
	if w.Code != 302 {
		t.Fatal("auth did not redirect")
	}
	u, _ := url.Parse(w.Header().Get("Location"))
	if u.Host != "github.com" || u.Query().Get("code_challenge_method") != "S256" || u.Query().Get("scope") != "public_repo" {
		t.Fatal("unexpected OAuth request")
	}
	c := w.Result().Cookies()[0]
	if !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
		t.Fatal("state cookie not protected")
	}
}
