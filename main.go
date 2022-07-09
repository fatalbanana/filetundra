package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatalbanana/filetundra/internal/env"
	"github.com/fatalbanana/filetundra/internal/idx"
	"github.com/fatalbanana/filetundra/internal/log"
	"github.com/fatalbanana/filetundra/internal/web"

	"go.uber.org/zap"
)

var (
	sigChan chan os.Signal
)

func main() {
	err := env.Process()
	if err != nil {
		panic(err)
	}

	err = log.SetupLogger()
	if err != nil {
		panic(err)
	}

	var ok bool

	defer func() {
		err := log.Logger.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error syncing logs: %s\n", err)
		}
		if !ok {
			os.Exit(1)
		}
	}()

	blugeDir, err := idx.GetBlugeDir()
	if err != nil {
		log.Logger.Error("failed to get bluge directory", zap.Error(err))
		ok = false
		return
	}

	ok = run(blugeDir)
}

func run(blugeDir string) bool {
	makeInitialIndex := false
	_, err := os.Stat(blugeDir)
	if err != nil {
		if os.IsNotExist(err) {
			makeInitialIndex = true
		} else {
			log.Logger.Error("failed to stat directory",
				zap.String("path", blugeDir), zap.Error(err))
			return false
		}
	}
	idx.Init(blugeDir)
	if makeInitialIndex {
		log.Logger.Info("performing initial indexing, please wait")
		err = idx.Initial()
		if err != nil {
			log.Logger.Error("failed to create index", zap.Error(err))
			return false
		}
		log.Logger.Info("created initial index")
	}

	go func() {
		err := web.RunWebserver()
		if err != nil && err != http.ErrServerClosed {
			log.Logger.Error("ListenAndServe error", zap.Error(err))
		}
	}()
	<-sigChan

	if web.Server == nil {
		return false
	}

	shutCtx, cancelShutCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutCtx()

	err = web.Server.Shutdown(shutCtx)
	if err != nil {
		log.Logger.Error("error shutting down webserver", zap.Error(err))
		err = web.Server.Close()
		if err != nil {
			log.Logger.Error("error closing webserver", zap.Error(err))
		}
		return false
	}
	return true
}

func init() {
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
}
