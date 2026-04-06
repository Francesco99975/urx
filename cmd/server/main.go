package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Francesco99975/urx/cmd/boot"
	"github.com/Francesco99975/urx/internal/cache"
	"github.com/Francesco99975/urx/internal/config"
	"github.com/Francesco99975/urx/internal/enums"

	"github.com/Francesco99975/urx/internal/database"

	"github.com/Francesco99975/urx/internal/helpers"
)

func main() {
	err := boot.LoadEnvVariables()
	if err != nil {
		panic(err)
	}

	var addr string
	if boot.Environment.GoEnv == enums.Environments.DEVELOPMENT {
		addr = "localhost:6379"
	} else if boot.Environment.GoEnv == enums.Environments.PRODUCTION {
		addr = "urxkeys:6379"
	}

	cache.InitCache(addr)

	if err := config.LoadManifest("./static"); err != nil {
		log.Fatalf("Failed to load Vite manifest: %v", err)
	}

	// Create a root ctx and a CancelFunc which can be used to cancel retentionMap goroutine
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	port := boot.Environment.Port

	database.Setup(boot.Environment.DSN)
	defer database.Close()

	e := createRouter(ctx)

	go func() {
		e.Logger.Infof("Running Server on port %s", port)
		e.Logger.Infof("Accessible locally at: http://localhost:%s", port)
		e.Logger.Infof("Accessible on the internet at: %s", boot.Environment.URL)
		e.Logger.Infof("Press Ctrl+C to stop the server and exit.")
		e.Logger.Fatal(e.Start(":" + port))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	helpers.Notify("urx", "Server is shutting down")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		helpers.Notify("urx", fmt.Sprintf("Server forced to shutdown: %v", err))
		e.Logger.Fatal(err)
	}
}
