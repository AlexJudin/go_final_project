package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/go_final_project/api"
	"github.com/AlexJudin/go_final_project/config"
	"github.com/AlexJudin/go_final_project/middleware"
	"github.com/AlexJudin/go_final_project/repository"
	"github.com/AlexJudin/go_final_project/usecases"
)

// @title Пользовательская документация API
// @description Итоговая работа по курсу "Go-разработчик с нуля" (Яндекс Практикум)
// @termsOfService spdante@mail.ru
// @contact.name Alexey Yudin
// @contact.email spdante@mail.ru
// @version 1.0.0
// @host localhost:7540
// @BasePath /
func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(cfg.LogLevel)

	db, err := repository.NewDB(cfg.DBFile)
	if err != nil {
		log.Fatalf("error connect to repository: %+v", err)
	}
	defer db.Close()

	// init repository
	repo := repository.NewNewRepository(db)

	// init usecases
	taskUC := usecases.NewTaskUsecase(repo)
	taskHandler := api.NewTaskHandler(taskUC)

	// init middleware
	authMiddleware := middleware.New(cfg)

	// init auth
	authHandler := middleware.NewAuthHandler(cfg)

	//webDir := "./web"
	r := chi.NewRouter()
	/*
		fileServer := http.FileServer(http.Dir(webDir))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			if filepath.Ext(r.URL.Path) == ".css" {
				w.Header().Set("Content-Type", "text/css")
			}
			fileServer.ServeHTTP(w, r)
		})
	*/
	r.Post("/api/signin", authHandler.GetAuthByPassword)
	r.Get("/api/nextdate", taskHandler.GetNextDate)
	r.Post("/api/task", authMiddleware.Auth(taskHandler.CreateTask))
	r.Get("/api/tasks", authMiddleware.Auth(taskHandler.GetTasks))
	r.Get("/api/task", authMiddleware.Auth(taskHandler.GetTaskById))
	r.Put("/api/task", authMiddleware.Auth(taskHandler.UpdateTask))
	r.Post("/api/task/done", authMiddleware.Auth(taskHandler.MakeTaskDone))
	r.Delete("/api/task", authMiddleware.Auth(taskHandler.DeleteTask))

	log.Info("Start http server")

	serverAddress := fmt.Sprintf("localhost:%s", cfg.Port)
	serverErr := make(chan error)

	httpServer := &http.Server{
		Addr:    serverAddress,
		Handler: r,
	}

	go func() {
		log.Infof("Listening on %s", serverAddress)
		if err = httpServer.ListenAndServe(); err != nil {
			serverErr <- err
		}
		close(serverErr)
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stop:
		log.Info("Stop signal received. Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = httpServer.Shutdown(ctx); err != nil {
			log.Errorf("error terminating server: %+v", err)
		}
		log.Info("The server has been stopped successfully")
	case err = <-serverErr:
		log.Errorf("server error: %+v", err)
	}
}
