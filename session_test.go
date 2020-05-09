package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func initMem(store IStorage) {

	Init(store, "TEST")
}

func TestMemSessions(t *testing.T) {
	store := NewMemSessionStorage()
	initMem(store)
	newSession := NewSession("/", "localhost", 300, true, false)

	newSession.Set("test", "test111")

	getValue := newSession.Get("test")
	t.Log("val:", getValue)

	l := newSession.Len()
	t.Log("len:", l)

	newSession.Del("test")

	l = newSession.Len()
	t.Log("len:", l)

	isGc := newSession.GC()
	t.Log("Gc:", isGc)

	id := newSession.ID
	name := cookieSessionName

	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "localhost:8999", nil)

	newSession.Save(resp, nil)

	req.AddCookie(&http.Cookie{
		Name:     name,
		Value:    id,
		Path:     "/",
		Domain:   "localhost",
		HttpOnly: false,
		Secure:   false,
	})

	sess, err := GetSession(req)
	if err == nil {
		t.Log("getSess:", sess)
	} else {
		t.Error(err.Error())
	}

	DelSession(resp, newSession)

}

func TestFileSessions(t *testing.T) {
	store := NewFileSessionStorage("./", "sess_")
	initMem(store)
	CunstomSessionStorage(store)
	SetCookieSessionName("TEST_SESSION")
	newSession := NewSession("/", "localhost", 300, false, false)

	getValue := newSession.Get("test")
	t.Log("val:", getValue)

	l := newSession.Len()
	t.Log("len:", l)

	newSession.Del("test")

	l = newSession.Len()
	t.Log("len:", l)

	isGc := newSession.GC()
	t.Log("Gc:", isGc)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "localhost:8999", nil)
	newSession.Save(resp, nil)

	id := newSession.ID
	name := cookieSessionName

	req.AddCookie(&http.Cookie{
		Name:     name,
		Value:    id,
		Path:     "/",
		Domain:   "localhost",
		HttpOnly: false,
		Secure:   false,
	})

	sess, err := GetSession(req)
	if err == nil {
		t.Log("getSess:", sess)
	} else {
		t.Error(err.Error())
	}

	DelSession(resp, newSession)
}

func TestRedisSessions(t *testing.T) {
	store, err := NewRedisSessionStorage("tcp", "127.0.0.1:6379")
	if err != nil {
		t.Error(err)
	}

	initMem(store)
	newSession := NewSession("/", "localhost", 300, true, false)

	newSession.Set("test", "test111")

	getValue := newSession.Get("test")
	t.Log("val:", getValue)

	l := newSession.Len()
	t.Log("len:", l)

	newSession.Del("test")

	l = newSession.Len()
	t.Log("len:", l)

	isGc := newSession.GC()
	t.Log("Gc:", isGc)

	id := newSession.ID
	name := cookieSessionName

	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "localhost:8999", nil)

	newSession.Save(resp, nil)

	req.AddCookie(&http.Cookie{
		Name:     name,
		Value:    id,
		Path:     "/",
		Domain:   "localhost",
		HttpOnly: false,
		Secure:   false,
	})

	sess, err := GetSession(req)
	if err == nil {
		t.Log("getSess:", sess)
	} else {
		t.Error(err.Error())
	}

	DelSession(resp, newSession)
}
