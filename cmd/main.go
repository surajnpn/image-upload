package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"image-upload/internal/api"
	"image-upload/internal/images"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ListenOn   int    `envconfig:"listen_on" default:"8080"`
	UploadPath string `envconfig:"upload_path" default:"./tmp_files"`
}

func main() {
	cfg, err := loadFromEnv()
	if err != nil {
		log.Fatal("error reading configs")
	}

	logger := log.Default()
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	imgService, err := images.New(cfg.UploadPath)
	if err != nil {
		log.Fatal("error with upload directory")
	}

	apiService := api.New(logger, imgService)
	httpRouter := setupHttpRouter(apiService)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.ListenOn),
		Handler: httpRouter,
	}
	go startHttpServer(srv, logger)

	<-ctx.Done()

	logger.Println("shutdown start")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("HTTP server forcefully shutdown ", err)
	}

	log.Println("shutdown complete")
}

func setupHttpRouter(api *api.APIService) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.MaxMultipartMemory = 4 << 20
	api.BindAPI(router.Group("api/v1/"))
	return router
}

func startHttpServer(srv *http.Server, log *log.Logger) {
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("error starting server ", err)
	}
}

func loadFromEnv() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
