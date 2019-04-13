package main

import (
	"github.com/CatWantsMeow/oml/server/core"
	"github.com/CatWantsMeow/oml/server/handlers"
	"net/http"
)

func main() {
	app := core.NewApp()
	app.HandlerFunc("/search/", app.LoggingMiddleware(handlers.HandleSearch))
	app.HandlerFunc("/reindex/", app.LoggingMiddleware(handlers.HandleReindex))
	app.HandlerFunc("/info/", app.LoggingMiddleware(handlers.HandleInfo))
	app.HandlerFunc("/", app.LoggingMiddleware(handlers.HandleServeFile))
	http.ListenAndServe(":"+app.AppPort, app)
}
