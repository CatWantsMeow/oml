package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/CatWantsMeow/oml/lib"
	"gopkg.in/olivere/elastic.v6"
	"html/template"
	"io/ioutil"
	"path"
	"strings"
)

const (
	url     = "http://localhost:9200"
	index   = "oml"
	mapping = `
    {
	    "mappings": {
		    "doc": {
			    "properties": {
				    "file": {
					    "type": "text"
				    },
				    "content": {
					    "type": "text",
                        "analyzer": "russian"
				    }
			    }
		    }
	    }
    }`
	lightPreTag  = `<span class="highlighted">`
	lightPostTag = `</span>`
)

type IndexEntry struct {
	File    string `json:"file"`
	Content string `json:"content"`
}

func NewIndexEntry(filename string, tag *lib.Tag) *IndexEntry {
	entry := &IndexEntry{File: filename}
	entry.build(tag)
	return entry
}

func (ie *IndexEntry) build(tag *lib.Tag) {
	ie.Content += string(lib.CompressSpaces(tag.Content)) + " "
	for _, tag := range tag.Nested {
		ie.build(tag)
	}
}

type SearchResult struct {
	File           string
	Score          string
	ContentPreview []template.HTML
}

type Indexer struct {
	App    *App
	Client *elastic.Client
}

func NewIndexer(app *App) *Indexer {
	idx := &Indexer{App: app}
	idx.connect()
	idx.initIndex()
	return idx
}

func (idx *Indexer) connect() {
	ctx := context.Background()

	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(false))
	if err != nil {
		idx.App.Logger.Fatal("Failed to connect to Elasticsearch: %s", err.Error())
	}

	info, code, err := client.Ping(url).Do(ctx)
	if err != nil {
		idx.App.Logger.Fatal("Failed to ping Elasticsearch: %s", err.Error())
	}
	idx.App.Logger.Info("Elasticsearch returned %d code and %s version", code, info.Version.Number)

	idx.Client = client
}

func (idx *Indexer) initIndex() {
	ctx := context.Background()

	exists, _ := idx.Client.IndexExists(index).Do(ctx)
	if !exists {
		_, err := idx.Client.CreateIndex(index).BodyString(mapping).Do(ctx)
		if err != nil {
			idx.App.Logger.Fatal("Failed to create es ReindexPages: %s", err.Error())
		}
	}
}

func (idx *Indexer) DeleteAll() {
	ctx := context.Background()
	qry := elastic.NewMatchAllQuery()
	_, err := idx.Client.DeleteByQuery(index).Type("doc").Query(qry).Do(ctx)
	if err != nil {
		idx.App.Logger.Warning("Failed to delete old entries: %s", err.Error())
	}
}

func (idx *Indexer) Search(phrase string) []*SearchResult {
	qry := elastic.NewMatchQuery("content", phrase)
	light := elastic.NewHighlight().Field("content").PreTags(lightPreTag).PostTags(lightPostTag)

	ctx := context.Background()
	result, err := idx.Client.Search(index).Type("doc").Query(qry).Highlight(light).Do(ctx)
	if err != nil {
		idx.App.Logger.Warning("Failed to perform search query: %s", err.Error())
		return make([]*SearchResult, 0)
	}

	results := make([]*SearchResult, 0)
	for _, hit := range result.Hits.Hits {
		entry := new(IndexEntry)
		err := json.Unmarshal(*hit.Source, entry)
		if err != nil {
			idx.App.Logger.Warning("Failed to parse search results: %s", err.Error())
			continue
		}

		preview := make([]template.HTML, 0)
		for _, light := range hit.Highlight["content"] {
			preview = append(preview, template.HTML(light+"..."))
		}

		result := &SearchResult{
			File:           entry.File,
			ContentPreview: preview,
			Score:          fmt.Sprintf("%.2f", *hit.Score),
		}
		results = append(results, result)
	}
	return results
}

func (idx *Indexer) IndexDir(dirPath string, prefix string) {
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		idx.App.Logger.Error(err.Error())
		return
	}

	ctx := context.Background()

	for _, f := range dir {
		filename := path.Join(dirPath, f.Name())

		if f.IsDir() {
			idx.IndexDir(filename, path.Join(prefix, f.Name()))
		}

		if !strings.HasPrefix(filename, ".") && strings.HasSuffix(filename, ".oml") {
			content, err := lib.ReadFile(filename)
			if err != nil {
				idx.App.Logger.Error("Failed to open %s: %s", filename, err.Error())
				continue
			}

			tag, err := lib.Parse(content)
			if err != nil {
				msg := strings.Replace(err.Error(), "\n", "; ", -1)
				idx.App.Logger.Error("Failed to parse %s: %s", filename, msg)
				continue
			}

			entry := NewIndexEntry(path.Join(prefix, f.Name()), tag)
			_, err = idx.Client.Index().Index(index).Type("doc").BodyJson(entry).Do(ctx)
			if err != nil {
				idx.App.Logger.Error("Failed to ReindexPages %s: %s", filename, err.Error())
			}

			idx.App.Logger.Info("Indexed %s", filename)
		}
	}
}
