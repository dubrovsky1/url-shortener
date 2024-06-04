package app

import (
	"context"
	"errors"
	"github.com/dubrovsky1/url-shortener/internal/config"
	"github.com/dubrovsky1/url-shortener/internal/handlers/api/shorten"
	"github.com/dubrovsky1/url-shortener/internal/handlers/api/user"
	"github.com/dubrovsky1/url-shortener/internal/handlers/geturl"
	"github.com/dubrovsky1/url-shortener/internal/handlers/ping"
	"github.com/dubrovsky1/url-shortener/internal/handlers/saveurl"
	"github.com/dubrovsky1/url-shortener/internal/middleware/auth"
	"github.com/dubrovsky1/url-shortener/internal/middleware/gzip"
	"github.com/dubrovsky1/url-shortener/internal/middleware/logger"
	"github.com/dubrovsky1/url-shortener/internal/service"
	"github.com/dubrovsky1/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	Flags   config.Config
	Storage storage.Storager
	Service *service.Service
}

func New() *App {
	flags := config.ParseFlags()

	stor, err := storage.GetStorage(flags)
	if err != nil {
		log.Fatal("Get storage error. ", err)
	}

	//создаем объект стоя бизнес-логики, который взаимодействует с базой
	serv := service.New(stor, 10, time.Second*10)

	return &App{
		Flags:   flags,
		Storage: stor,
		Service: serv,
	}
}

func (a *App) Run() {
	r := chi.NewRouter()
	r.Post("/", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(saveurl.SaveURL(a.Service, a.Flags.ResultShortURL)))))
	r.Post("/api/shorten", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(shorten.Shorten(a.Service, a.Flags.ResultShortURL)))))
	r.Post("/api/shorten/batch", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(shorten.Batch(a.Service)))))
	r.Get("/{id}", logger.WithLogging(gzip.GzipMiddleware(geturl.GetURL(a.Service))))
	r.Get("/ping", logger.WithLogging(gzip.GzipMiddleware(ping.Ping(a.Flags.ConnectionString))))
	r.Get("/api/user/urls", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(user.ListByUserID(a.Service)))))
	r.Delete("/api/user/urls", auth.Auth(logger.WithLogging(gzip.GzipMiddleware(user.DeleteURL(a.Service)))))

	serv := http.Server{
		Addr:    a.Flags.Host,
		Handler: r,
	}

	go func() {
		if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen and serve returned err: %v", err)
		}
	}()
	logger.Sugar.Infow("Server is listening", "host", a.Flags.Host)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go a.RunPprof(ctx)

	//запуск удаления записей
	a.Service.Run(ctx)

	<-ctx.Done()

	if err := serv.Shutdown(ctx); err != nil {
		logger.Sugar.Infow("Server shutdown error", "err", err.Error())
	}

	a.Close()

	logger.Sugar.Infow("Shutting down server gracefully")
}

func (a *App) Close() {
	a.Storage.Close()
	logger.Sugar.Infow("Storage closed")

	a.Service.Close()
	logger.Sugar.Infow("Service closed")
}

func (a *App) RunPprof(ctx context.Context) {
	r := chi.NewRouter()

	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	//r.HandleFunc("/debug/pprof/{cmd}", pprof.Index)

	r.Handle("/debug/pprof/block", pprof.Handler("block"))
	r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	pprofServ := http.Server{
		Addr:    "localhost:9090",
		Handler: r,
	}

	go func() {
		_ = pprofServ.ListenAndServe()
	}()
	logger.Sugar.Infow("Server pprof is listening", "host", "localhost:9090")

	<-ctx.Done()
	pprofServ.SetKeepAlivesEnabled(false)
	_ = pprofServ.Shutdown(ctx)

	logger.Sugar.Infow("Shutting down pprof server gracefully")
}
