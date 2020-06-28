package handler

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/rdnply/wschat/cmd/socket"
	"github.com/rdnply/wschat/internal/ehttp"
	"github.com/rdnply/wschat/internal/inmemory"
	"github.com/rdnply/wschat/internal/user"
	"github.com/rdnply/wschat/pkg/log/logger"
	"html/template"
	"net/http"
)

type Handler struct {
	hub       *socket.Hub
	templates *Templates
	//users     map[string]bool
	userStorage user.Storage
	logger      logger.Logger
}

func New(hub *socket.Hub, log logger.Logger) (*Handler, error) {
	return &Handler{
		hub:         hub,
		templates:   readTemplates(),
		userStorage: inmemory.NewUserStorage(),
		logger:      log,
	}, nil
}

type Templates struct {
	login *template.Template
	chat  *template.Template
}

func readTemplates() *Templates {
	return &Templates{
		login: readTemplate("login"),
		chat:  readTemplate("chat"),
	}
}

func readTemplate(name string) *template.Template {
	path := "./static/templates/"

	t := template.Must(template.New(name + ".html").
		ParseGlob(path + name + ".html"))

	return t
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/login", errorHandling(h.loginForm, h.logger))
		r.Post("/login", errorHandling(h.register, h.logger))
		r.Get("/chat", errorHandling(h.chat, h.logger))
	})

	r.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		socket.ServeWS(h.hub, w, req)
	})
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return r
}

type handlerFunc func(http.ResponseWriter, *http.Request) error

func errorHandling(h handlerFunc, l logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			e, ok := err.(ehttp.HTTPError)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if e.Detail != "" {
				l.Errorf(e.Detail)
			}

			w.WriteHeader(e.StatusCode)

			if e.Msg != "" {
				w.Header().Add("Content-Type", "application/json")

				out, err := json.Marshal(e)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// no need to handle error here
				_, _ = w.Write(out)
			}
		}
	}
}
