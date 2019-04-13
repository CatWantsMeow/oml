package core

import (
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Response struct {
	http.ResponseWriter
	Status int
}

func (r *Response) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

type HTTPError struct {
	Status  int
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func NewHTTPError(status int, error string, details string) *HTTPError {
	return &HTTPError{Status: status, Error: error, Details: details}
}

type HandlerFunc func(*App, *http.Request, *Response)

type App struct {
	AppPort  string
	ServeDir string

	CacheTemplates bool
	TemplatePath   string
	templateCache  map[string]*template.Template

	Mux    *http.ServeMux
	Logger *Logger
	Idx    *Indexer
}

func NewApp() *App {
	app := App{
		Mux:            http.NewServeMux(),
		Logger:         NewLogger(LevelDebug, os.Stdout),
		CacheTemplates: false,
		templateCache:  make(map[string]*template.Template),
	}
	app.loadOptions()

	app.Idx = NewIndexer(&app)
	app.ReindexPages()

	app.Logger.Info("Serving files from %s on %s port", app.ServeDir, app.AppPort)
	return &app
}

func (app *App) loadOptions() {
	app.AppPort = app.getOption("APP_PORT", "8080")

	matches, err := filepath.Glob("**/templates/main.html")
	if err != nil || matches == nil || len(matches) == 0 {
		app.Logger.Fatal("Templates path was not found.")
	}
	app.TemplatePath = path.Dir(matches[0])

	level := app.getOption("LOG_LEVEL", "DEBUG")
	err = app.Logger.SetLevel(level)
	if err != nil {
		app.Logger.Fatal("Unknown log level '%s'", level)
	}

	dir, err := filepath.Abs(app.getOption("SERVE_DIR", ""))
	if err != nil {
		app.Logger.Fatal(err.Error())
	} else if _, err := os.Stat(dir); os.IsNotExist(err) {
		app.Logger.Fatal("Path %s does not exist", app.ServeDir)
	}
	app.ServeDir = dir
}

func (app *App) ReindexPages() {
	app.Logger.Info("Indexing files from %s", app.ServeDir)
	app.Idx.DeleteAll()
	app.Idx.IndexDir(app.ServeDir, "")
}

func (app *App) GetPageInfo() ([]*Node, [][]bool, [][]int) {
	crawler := NewCrawler(app)
	nodes := crawler.Crawl(app.ServeDir, "")
	a := crawler.BuildAdjacencyMatrix(nodes)
	w := crawler.CountShortestPaths(nodes)
	return nodes, a, w
}

func (app *App) getOption(option string, def string) string {
	val := os.Getenv(option)
	if val == "" {
		if def != "" {
			app.Logger.Warning("Env variable %s is not defined.", option)
			return def
		}
		app.Logger.Fatal("Env variable %s must be defined.", option)
	}
	return val
}

func (app *App) LoadTemplate(filename string) *template.Template {
	load := func(filename string) *template.Template {
		t, err := template.ParseFiles(
			path.Join(app.TemplatePath, "main.html"),
			path.Join(app.TemplatePath, filename),
		)
		if err != nil {
			app.Logger.Fatal(err.Error())
		}
		return t
	}

	if app.CacheTemplates {
		if _, ok := app.templateCache[filename]; !ok {
			app.templateCache[filename] = load(filename)
		}
		return app.templateCache[filename]
	}
	return load(filename)
}

func (app *App) WriteHttpError(rsp *Response, error *HTTPError) {
	rsp.WriteHeader(error.Status)
	err := app.LoadTemplate("error.html").ExecuteTemplate(rsp, "main", error)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
		rsp.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *App) AppContextMiddleware(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		rsp := &Response{ResponseWriter: w, Status: 200}
		handler(app, req, rsp)
	}
}

func (app *App) LoggingMiddleware(handler HandlerFunc) HandlerFunc {
	return func(app *App, req *http.Request, rsp *Response) {
		started := time.Now()
		handler(app, req, rsp)

		dur := time.Since(started).Seconds() * 1000
		app.Logger.Info(
			"%s %s - %d in %.2f ms",
			req.Method, req.URL.Path, rsp.Status, dur,
		)
	}
}

func (app *App) HandlerFunc(pattern string, handlerFunc HandlerFunc) {
	app.Mux.HandleFunc(pattern, app.AppContextMiddleware(handlerFunc))
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.Mux.ServeHTTP(w, r)
}
