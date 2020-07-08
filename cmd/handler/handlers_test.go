package handler

import (
	"bytes"
	"github.com/rdnply/wschat/cmd/wssocket"
	"github.com/rdnply/wschat/internal/message"
	"github.com/rdnply/wschat/internal/user"
	"github.com/rdnply/wschat/pkg/log/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type mockUserStorage struct {
	s user.Storage
	u user.User
}

func (m mockUserStorage) Find(login string) bool {
	return m.u.Login == login
}

func (m mockUserStorage) Add(login string) {}

func (m mockUserStorage) FindMessages(login string, companion string) []*message.Message {
	msgs, ok := m.u.Messages[companion]
	if !ok {
		return nil
	}

	return msgs
}

func (m mockUserStorage) AddMessage(login string, companion string, message *message.Message) {}

func (m mockUserStorage) GetLogins() []string {
	return []string{m.u.Login}
}

type mockLogger struct {
	logger.Logger
}

func (m mockLogger) Debugf(format string, args ...interface{}) {}
func (m mockLogger) Infof(format string, args ...interface{})  {}
func (m mockLogger) Warnf(format string, args ...interface{})  {}
func (m mockLogger) Errorf(format string, args ...interface{}) {}
func (m mockLogger) Fatalf(format string, args ...interface{}) {}
func (m mockLogger) Panicf(format string, args ...interface{}) {}

func init() {
	os.Chdir("../..") // change to directory that contains main.go
}

func TestLoginForm(t *testing.T) {
	req, err := http.NewRequest("GET", "/login", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	mockUserStorage := new(mockUserStorage)
	hub := wssocket.NewHub(mockUserStorage)
	mockLogger := new(mockLogger)

	h := New(hub, mockUserStorage, mockLogger)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(errorHandling(h.loginForm, mockLogger))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("loginForm handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}
}

func TestRegisterLoginAlreadyExist(t *testing.T) {
	userLogin := "THIS LOGIN IS ALREADY EXIST"
	req, err := http.NewRequest("POST", "/login", bytes.NewBufferString("login="+userLogin))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockUserStorage := new(mockUserStorage)
	hub := wssocket.NewHub(mockUserStorage)
	mockLogger := new(mockLogger)

	u := user.User{Login: userLogin}
	mockUserStorage.u = u

	h := New(hub, mockUserStorage, mockLogger)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(errorHandling(h.register, mockLogger))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("register handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	var expected bytes.Buffer
	tmpl := h.templates.login

	err = tmpl.Execute(&expected, struct {
		Login string
		Error string
	}{
		Login: userLogin,
		Error: "This login is already exist",
	})
	if err != nil {
		t.Fatalf("can't execute template")
	}

	if !bytes.Equal(rr.Body.Bytes(), expected.Bytes()) {
		t.Errorf("register handler returned wrong body: got\n %v\n, want\n %v\n",
			rr.Body.String(), expected.String())
	}
}

func TestRegisterRedirectionToChat(t *testing.T) {
	userLogin := "THIS LOGIN IS FREE"
	req, err := http.NewRequest("POST", "/login", bytes.NewBufferString("login="+userLogin))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockUserStorage := new(mockUserStorage)
	hub := wssocket.NewHub(mockUserStorage)
	mockLogger := new(mockLogger)

	u := user.User{Login: "LOGIN OF ANOTHER USER"}
	mockUserStorage.u = u

	h := New(hub, mockUserStorage, mockLogger)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(errorHandling(h.register, mockLogger))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("register handler returned wrong status code: got %v, want %v",
			status, http.StatusSeeOther)
	}
}

func TestChatCorrect(t *testing.T) {
	userLogin := "THIS LOGIN IS FREE"
	req, err := http.NewRequest("GET", "/chat", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	q := req.URL.Query()
	q.Add("login", userLogin)
	req.URL.RawQuery = q.Encode()

	mockUserStorage := new(mockUserStorage)
	hub := wssocket.NewHub(mockUserStorage)
	mockLogger := new(mockLogger)

	u := user.User{Login: "LOGIN OF ANOTHER USER"}
	mockUserStorage.u = u

	h := New(hub, mockUserStorage, mockLogger)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(errorHandling(h.chat, mockLogger))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("chat handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	var expected bytes.Buffer
	tmpl := h.templates.chat

	err = tmpl.Execute(&expected, mockUserStorage.GetLogins())
	if err != nil {
		t.Fatalf("can't execute template")
	}

	if !bytes.Equal(rr.Body.Bytes(), expected.Bytes()) {
		t.Errorf("chat handler returned wrong body: got\n %v\n, want\n %v\n",
			rr.Body.String(), expected.String())
	}
}

func TestChatEmptyLogin(t *testing.T) {
	userLogin := "" // empty string for login
	req, err := http.NewRequest("GET", "/chat", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	q := req.URL.Query()
	q.Add("login", userLogin)
	req.URL.RawQuery = q.Encode()

	mockUserStorage := new(mockUserStorage)
	hub := wssocket.NewHub(mockUserStorage)
	mockLogger := new(mockLogger)

	u := user.User{Login: "LOGIN OF ANOTHER USER"}
	mockUserStorage.u = u

	h := New(hub, mockUserStorage, mockLogger)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(errorHandling(h.chat, mockLogger))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("chat handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}
}

func TestGetMessagesCorrect(t *testing.T) {
	userLogin := "ANY LOGIN"
	req, err := http.NewRequest("GET", "/chat/messages", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	companionLogin := "COMPANION WITH WHOM EXIST MESSAGES"
	q := req.URL.Query()
	q.Add("login", userLogin)
	q.Add("companion", companionLogin)
	req.URL.RawQuery = q.Encode()

	mockUserStorage := new(mockUserStorage)
	hub := wssocket.NewHub(mockUserStorage)
	mockLogger := new(mockLogger)

	messages := make(map[string][]*message.Message)
	msg := &message.Message{From: userLogin, To: companionLogin, Message: "MESSAGE"}
	messages[companionLogin] = append(messages[companionLogin], msg)

	u := user.User{Login: "LOGIN OF ANOTHER USER", Messages: messages}
	mockUserStorage.u = u

	h := New(hub, mockUserStorage, mockLogger)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(errorHandling(h.getMessages, mockLogger))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("getMessages handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `[{"from":"` + userLogin + `","to":"` + companionLogin + `","message":"MESSAGE"}]`

	if rr.Body.String() != expected {
		t.Errorf("getMessages handler returned wrong json: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetMessagesEmptyQueryParams(t *testing.T) {
	userLogin := ""
	req, err := http.NewRequest("GET", "/chat/messages", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	companionLogin := ""
	q := req.URL.Query()
	q.Add("login", userLogin)
	q.Add("companion", companionLogin)
	req.URL.RawQuery = q.Encode()

	mockUserStorage := new(mockUserStorage)
	hub := wssocket.NewHub(mockUserStorage)
	mockLogger := new(mockLogger)

	h := New(hub, mockUserStorage, mockLogger)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(errorHandling(h.getMessages, mockLogger))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("getMessages handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}
}
