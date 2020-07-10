package handler

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rdnply/wschat/internal/errorhttp"
	"html/template"
	"io"
	"net/http"
	"strings"
)

func (h *Handler) loginForm(w http.ResponseWriter, r *http.Request) error {
	return renderTemplate(w, h.templates.login, nil)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) error {
	type invalidLogin struct{
		Login string
		Error string
	}

	const MinNumSymbolsInLogin = 4

	login := r.PostFormValue("login")
	if strings.ContainsRune(login, ' ') || len(login) < MinNumSymbolsInLogin {
		return renderTemplate(w, h.templates.login, invalidLogin{
			Login: login,
			Error: "Login must not contain spaces and be less than 4 characters",
		})
	} else if h.userStorage.Find(login) {
		return renderTemplate(w, h.templates.login, invalidLogin{
			Login: login,
			Error: "This login is already exist",
		})
	}

	http.Redirect(w, r, "/chat?login="+login, http.StatusSeeOther)

	return nil
}

func (h *Handler) chat(w http.ResponseWriter, r *http.Request) error {
	login := r.URL.Query().Get("login")
	if login == "" {
		http.Error(w, "can't find login in url params", http.StatusBadRequest)
	}

	defer h.userStorage.Add(login)
	return renderTemplate(w, h.templates.chat, h.userStorage.GetLogins())
}

func (h *Handler) getMessages(w http.ResponseWriter, r *http.Request) error {
	login := r.URL.Query().Get("login")
	companion := r.URL.Query().Get("companion")
	if login == "" || companion == "" {
		http.Error(w, "can't find login in url params", http.StatusBadRequest)
	}

	messages := h.userStorage.FindMessages(login, companion)

	err := respondJSON(w, messages)
	if err != nil {
		detail := fmt.Sprintf("can't make json respond with messages: %v", err)
		return errorhttp.InternalServerErr(detail)
	}

	return nil
}

func renderTemplate(w io.Writer, tmpl *template.Template, payload interface{}) error {
	err := tmpl.Execute(w, payload)
	if err != nil {
		detail := fmt.Sprintf("can't execute template with name: %v: %v", tmpl.Name(), err)
		return errorhttp.InternalServerErr(detail)
	}

	return nil
}

func respondJSON(w http.ResponseWriter, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrapf(err, "can't marshal respond to json")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	c, err := w.Write(response)
	if err != nil {
		msg := fmt.Sprintf("can't write json data in respond, code: %v", c)
		return errors.Wrapf(err, msg)
	}

	return nil
}
