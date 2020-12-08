package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mercadolibre.com/di/practice/business"
	"mercadolibre.com/di/practice/handlers"
	"mercadolibre.com/di/practice/internal"
	"mercadolibre.com/di/practice/weather"
)

var (
	port            = internal.GetEnv("PORT", "3001")
	weatherProvider = internal.GetEnv("WEATHER_PROVIDER", "weather-bit")
)

func main() {

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      loadDependencies(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Printf("Starting HTTP Server. Listening at %q", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("%v", err)
		} else {
			log.Println("Server closed!")
		}
	}()

	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, os.Interrupt, syscall.SIGTERM)
	sig := <-sigquit
	log.Printf("caught sig: %+v", sig)
	log.Printf("Gracefully shutting down server...")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("Unable to shut down server: %v", err)
	} else {
		log.Println("Server stopped")
	}
}

func loadDependencies() http.Handler {
	wProvider, err := weather.GetProvider(weatherProvider)
	if err != nil {
		panic(err)
	}

	fService, err := weather.NewForecastService(wProvider)
	if err != nil {
		panic(err)
	}

	bfCalculator := business.NewBeerForecast(fService)

	bfResolver := handlers.NewBeerForecastResolver(bfCalculator)

	return NewRouter(bfResolver)
}
