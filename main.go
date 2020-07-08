package main

import (
	"context"
	"flag"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rdnply/wschat/cmd/handler"
	"github.com/rdnply/wschat/cmd/wssocket"
	"github.com/rdnply/wschat/internal/inmemory"
	"github.com/rdnply/wschat/pkg/log/logger"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var port = flag.String("port", "5000", "the port which server listen")

	flag.Parse()

	logger := initLogger()

	userStorage := inmemory.NewUserStorage()

	hub := wssocket.NewHub(userStorage)

	go hub.Run()

	h := handler.New(hub, userStorage, logger)

	srv := initServer(h, "", *port)

	const Duration = 5
	go gracefulShutdown(srv, Duration*time.Second, logger)

	logger.Infof("Server is running at %s", *port)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func initServer(h *handler.Handler, host string, port string) *http.Server {
	r := routes(h)
	addr := net.JoinHostPort(host, port)
	srv := &http.Server{Addr: addr, Handler: r}

	return srv
}

func routes(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	const Duration = 60

	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(Duration * time.Second))

	r.Mount("/", h.Routes())

	return r
}

func gracefulShutdown(srv *http.Server, timeout time.Duration, logger logger.Logger) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Infof("Shutting down server with %s timeout", timeout)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("could not shutdown server:%v", err)
	}
}

func initLogger() logger.Logger {
	config := logger.Configuration{
		EnableConsole:     true,
		ConsoleLevel:      logger.Debug,
		ConsoleJSONFormat: true,
		EnableFile:        true,
		FileLevel:         logger.Info,
		FileJSONFormat:    true,
		FileLocation:      "log.log",
	}

	logger, err := logger.New(config, logger.InstanceZapLogger)
	if err != nil {
		log.Fatal("could not instantiate logger: ", err)
	}

	return logger
}
