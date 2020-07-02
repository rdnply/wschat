package handler

import (
	"fmt"
	"github.com/rdnply/wschat/internal/ehttp"
	"html/template"
	"io"
	"net/http"
)

func (h *Handler) loginForm(w http.ResponseWriter, r *http.Request) error {
	return renderTemplate(w, h.templates.login, nil)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) error {
	login := r.PostFormValue("login")
	if h.userStorage.Find(login) {
		return renderTemplate(w, h.templates.login, struct {
			Login string
			Error string
		}{
			Login: login,
			Error: "That login is already exist",
		})
	}

	//h.userStorage.Add(login)
	http.Redirect(w, r, "/chat?login="+login, http.StatusSeeOther)

	return nil
}

func (h *Handler) chat(w http.ResponseWriter, r *http.Request) error {
	login := r.URL.Query().Get("login")
	if login == "" {
		http.Error(w, "can't find login in url params", http.StatusBadRequest)
	}

	defer h.userStorage.Add(login)
	return renderTemplate(w, h.templates.chat, h.userStorage)
}

func renderTemplate(w io.Writer, tmpl *template.Template, payload interface{}) error {
	err := tmpl.Execute(w, payload)
	if err != nil {
		detail := fmt.Sprintf("can't execute template with name: %v: %v", tmpl.Name(), err)
		return ehttp.InternalServerErr(detail)
	}

	return nil
}
