package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"dsoob/backend/core"
	"dsoob/backend/tools"
)

func main() {
	time.Local = time.UTC

	// Debug Commands
	// 	We exit immediately afterwards because the server starting up afterwards
	// 	is most likely isn't what you want, and it probably leaked memory elsewhere.
	for _, str := range os.Args {
		if strings.EqualFold(str, "debug_database_update_geolocation") {
			core.DebugDatabaseUpdateGeolocation()
			return
		}
		if strings.EqualFold(str, "debug_email_render_templates") {
			core.DebugEmailRenderTemplates()
			return
		}
		if strings.EqualFold(str, "debug_image_resizer") {
			core.DebugImageResizer()
			return
		}
	}

	// Startup Services
	// 	Logger are unique and must be started specifically,
	// 	everything else can be started at the same time
	var stopCtx, stop = context.WithCancel(context.Background())
	var stopWg sync.WaitGroup
	var syncWg sync.WaitGroup

	tools.LoggerMain.Log(tools.INFO, "Starting Services")
	for _, fn := range []func(stop context.Context, await *sync.WaitGroup){
		tools.GeolocateSetup,
		tools.DatabaseSetup,
	} {
		syncWg.Add(1)
		go func() {
			defer syncWg.Done()
			fn(stopCtx, &stopWg)
		}()
	}
	syncWg.Wait()
	go StartupHTTP(stopCtx, &stopWg)

	// Await Shutdown Signal
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-cancel
	stop()

	// Begin Shutdown Process
	timeout, finish := context.WithTimeout(context.Background(), tools.TIMEOUT_SHUTDOWN)
	defer finish()
	go func() {
		<-timeout.Done()
		if timeout.Err() == context.DeadlineExceeded {
			tools.LoggerMain.Log(tools.FATAL, "Shutdown Deadline Exceeded")
		}
	}()
	stopWg.Wait()
	os.Exit(0)
}

func StartupHTTP(stop context.Context, await *sync.WaitGroup) {

	// Optimized to prevent malicious attacks but shouldn't
	// really bother devices on slower networks :)

	svr := http.Server{
		Handler:           core.SetupMux(),
		Addr:              tools.HTTP_ADDRESS,
		MaxHeaderBytes:    4096,
		IdleTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       10 * time.Second,
	}
	if tools.HTTP_TLS_ENABLED {
		tls, err := tools.NewTLSConfig(
			tools.HTTP_TLS_CERT,
			tools.HTTP_TLS_KEY,
			tools.HTTP_TLS_CA,
		)
		if err != nil {
			tools.LoggerHTTP.Log(tools.FATAL, "TLS Configuration Error: %s", err)
			return
		}
		svr.TLSConfig = tls
	}

	// Shutdown Logic
	await.Add(1)
	go func() {
		defer await.Done()
		<-stop.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), tools.TIMEOUT_CONTEXT)
		defer cancel()

		if err := svr.Shutdown(shutdownCtx); err != nil {
			tools.LoggerHTTP.Log(tools.ERROR, "Shutdown error: %s", err)
		}

		tools.LoggerHTTP.Log(tools.INFO, "Closed")
	}()

	// Server Startup
	var err error
	tools.LoggerHTTP.Log(tools.INFO, "Listening @ %s", svr.Addr)
	if tools.HTTP_TLS_ENABLED {
		err = svr.ListenAndServeTLS("", "")
	} else {
		err = svr.ListenAndServe()
	}
	if err != http.ErrServerClosed {
		tools.LoggerHTTP.Log(tools.FATAL, "Startup Failed: %s", err)
	}
}
