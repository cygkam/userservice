package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-service/internal/api"
	"user-service/internal/config"
	"user-service/internal/db"
	"user-service/internal/service"

	"github.com/sirupsen/logrus"
)

func Start() {
	logrus.Info("Starting user service")

	config.Load()
	configureLogger()

	dbCfg := &db.DbConfig{
		Host:         config.GetString("psql.host"),
		Port:         config.GetString("psql.port"),
		Username:     config.GetString("psql.username"),
		DatabaseName: config.GetString("psql.db.name"),
		Password:     config.GetString("psql.password"),
		Timezone:     config.GetString("psql.timezone"),
	}

	_, err := db.MigrateSchema(context.Background(), dbCfg)
	if err != nil {
		logrus.Fatalf("Database migration failed: %v", err)
	}

	db, err := db.CreateDbConnection(context.Background(), dbCfg)
	if err != nil {
		logrus.Fatalf("Database connection initialization failed: %v", err)
	}

	us := service.NewUserService(db)
	go us.Cache.StartHTTPPoolServer()

	h := &api.Handler{
		UserHandler: &api.UserHandler{
			UserService: us,
		},
	}

	server := &http.Server{
		Addr:    ":" + config.GetString("config.port"),
		Handler: api.InitRouter(h),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()

	gracefulShutdown(server)
}

func configureLogger() {
	logLevel, err := logrus.ParseLevel(config.GetString("config.logging.level"))
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.SetLevel(logLevel)
}

func gracefulShutdown(httpServer *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Shutting down gracefully.")
}
