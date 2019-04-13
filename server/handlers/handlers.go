package handlers

import (
	"github.com/CatWantsMeow/oml/lib"
	"github.com/CatWantsMeow/oml/server/core"
	"html/template"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	indexPage = "index.oml"
)

func toHttpError(filePath string, err error, app *core.App) *core.HTTPError {
	if os.IsNotExist(err) || os.IsPermission(err) {
		app.Logger.Warning("File %s does not exist.", filePath)
		return core.NewHTTPError(http.StatusNotFound, "Not Found", "")
	}

	app.Logger.Error(err.Error())
	return core.NewHTTPError(
		http.StatusInternalServerError,
		"Internal Server Error",
		"",
	)
}

func HandleServeFile(app *core.App, req *http.Request, rsp *core.Response) {
	if req.URL.Path == "/" {
		_, err := os.Stat(path.Join(app.ServeDir, indexPage))
		if err == nil {
			http.Redirect(rsp, req, indexPage, http.StatusMovedPermanently)
		}
	}

	if strings.HasPrefix(req.URL.Path, "/$internal") {
		filePath := strings.Replace(req.URL.Path, "/$internal", "", 1)
		filePath = path.Join(app.TemplatePath, filePath)
		http.ServeFile(rsp, req, filePath)
		return
	}

	filePath := path.Join(app.ServeDir, req.URL.Path)
	content, err := lib.ReadFile(filePath)
	if err != nil {
		app.WriteHttpError(rsp, toHttpError(filePath, err, app))
		return
	}

	if !strings.HasSuffix(filePath, ".oml") {
		rsp.Write(content)
		return
	}

	tag, err := lib.Parse(content)
	if err != nil {
		err := core.NewHTTPError(
			http.StatusInternalServerError,
			"Internal Server Error",
			err.Error(),
		)
		app.WriteHttpError(rsp, err)
		return
	}

	params := struct{ Content template.HTML }{template.HTML(tag.HTML())}
	err = app.LoadTemplate("file.html").ExecuteTemplate(rsp, "main", &params)
	if err != nil {
		app.WriteHttpError(rsp, toHttpError(filePath, err, app))
	}
}

func HandleSearch(app *core.App, req *http.Request, rsp *core.Response) {
	results := make([]*core.SearchResult, 0)
	if req.Method == http.MethodPost {
		req.ParseForm()
		phrase, ok := req.Form["phrase"]
		if !ok || len(phrase) < 1 {
			err := core.NewHTTPError(
				http.StatusBadRequest,
				"Bad Request",
				"Request contains invalid form.",
			)
			app.WriteHttpError(rsp, err)
			return
		}
		results = app.Idx.Search(phrase[0])
	}

	params := struct{ Results []*core.SearchResult }{results}
	err := app.LoadTemplate("search.html").ExecuteTemplate(rsp, "main", &params)
	if err != nil {
		app.WriteHttpError(rsp, toHttpError("", err, app))
	}
}

func HandleInfo(app *core.App, req *http.Request, rsp *core.Response) {
	nodes, a, w := app.GetPageInfo()

	params := struct{ Nodes []*core.Node; A [][]bool; W [][]int }{Nodes: nodes, A: a, W: w}
	err := app.LoadTemplate("info.html").ExecuteTemplate(rsp, "main", &params)
	if err != nil {
		app.WriteHttpError(rsp, toHttpError("", err, app))
	}
}

func HandleReindex(app *core.App, req *http.Request, rsp *core.Response) {
	app.ReindexPages()
	http.Redirect(rsp, req, "/", http.StatusNotModified)
}
