package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"prodLoaderREST/internal/api"
	"prodLoaderREST/internal/broker"
	"prodLoaderREST/internal/config"
	"prodLoaderREST/internal/logger"
	"prodLoaderREST/internal/services/consumer/vk"
	"prodLoaderREST/internal/storage/sqlite"
	"syscall"
	"time"

	vkApi "github.com/SevereCloud/vksdk/v3/api"
	"github.com/gin-gonic/gin"
)

func main() {
	//ctx := context.Background()

	cfg := config.MustRead()

	log := logger.New(cfg.Log)

	log.Info("App is starting")

	gin.SetMode(gin.ReleaseMode)

	storage, err := sqlite.New(log, cfg.DbPath)
	if err != nil {
		log.Error("Failed to create storage", "err", err.Error())
		return
	}

	VK := vkApi.NewVK(cfg.VkToken)

	vkClient := vk.New(log, VK, storage, cfg.VkGroupID)

	log.Info("Autharizated vk:", "Name:", vkClient.GetClientName())

	API := api.New(log)
	API.Setup()

	go vkClient.Load(broker.VKProductChannel)

	srv := http.Server{
		Addr:    cfg.ServerHost + ":" + cfg.ServerPort,
		Handler: API.Router,
	}

	chanErrors := make(chan error, 1)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("HTTP server started", "Addres", srv.Addr)
		chanErrors <- srv.ListenAndServe()
	}()

	go func() {
		log.Info("Started to ping databse")
		for {
			time.Sleep(5 * time.Second)
			err := storage.Ping()
			if err != nil {
				chanErrors <- err
				break
			}
		}

	}()

	// gracefull shutdown
	select {
	case err := <-chanErrors:
		log.Error("Shutting down. Critical error:", "err", err)

		shutdown <- syscall.SIGTERM
	case sig := <-shutdown:
		log.Error("received signal, starting graceful shutdown", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("server graceful shutdown failed", "err", err)
			err = srv.Close()
			if err != nil {
				log.Error("forced shutdown failed", "err", err)
			}
		}

		storage.Close()

		log.Info("shutdown completed")

	}

}
