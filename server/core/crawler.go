package core

import (
	"github.com/CatWantsMeow/oml/lib"
	"io/ioutil"
	"path"
	"strings"
)

type Node struct {
	File  string
	Links []string
}

func NewNode(filename string, tag *lib.Tag) *Node {
	node := &Node{File: filename}
	node.build(tag)
	return node
}

func (n *Node) build(tag *lib.Tag) {
	if tag.Name == "link" {
		uri := lib.Get(tag.Options, "uri", "")
		uri = strings.Trim(uri, "/")

		if len(uri) > 0 && !lib.In(uri, n.Links) {
			n.Links = append(n.Links, uri)
		}
	}

	for _, tag := range tag.Nested {
		n.build(tag)
	}
}

type Crawler struct {
	App     *App
}

func NewCrawler(app *App) *Crawler {
	return &Crawler{
		App:     app,
	}
}

func (c *Crawler) Crawl(dirPath string, prefix string) []*Node {
	nodes := make([]*Node, 0)

	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		c.App.Logger.Error(err.Error())
		return nodes
	}

	for _, f := range dir {
		filename := path.Join(dirPath, f.Name())

		if f.IsDir() {
			nested := c.Crawl(filename, path.Join(prefix, f.Name()))
			nodes = append(nodes, nested...)
		}

		if !strings.HasPrefix(filename, ".") && strings.HasSuffix(filename, ".oml") {
			content, err := lib.ReadFile(filename)
			if err != nil {
				c.App.Logger.Error("Failed to open %s: %s", filename, err.Error())
				continue
			}

			tag, err := lib.Parse(content)
			if err != nil {
				msg := strings.Replace(err.Error(), "\n", "; ", -1)
				c.App.Logger.Error("Failed to parse %s: %s", filename, msg)
				continue
			}

			node := NewNode(path.Join(prefix, f.Name()), tag)
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (c *Crawler) BuildAdjacencyMatrix(nodes []*Node) [][]bool {
	indices := make(map[string]int)
	for i, node := range nodes {
		indices[node.File] = i
	}

	a := make([][]bool, len(indices), len(indices))
	for i, node := range nodes {
		a[i] = make([]bool, len(indices), len(indices))
		for _, link := range node.Links {
			if _, ok := indices[link]; ok {
				a[i][indices[link]] = true
			}
		}
	}
	return a
}

func (c *Crawler) CountShortestPaths(nodes []*Node) [][]int {
	a := c.BuildAdjacencyMatrix(nodes)

	w := make([][]int, len(a), len(a))
	for i := 0; i < len(a); i++ {
		w[i] = make([]int, len(a[i]), len(a[i]))
		for j := 0; j < len(a[i]); j++ {
			if a[i][j] {
				w[i][j] = 1
			} else {
				w[i][j] = 1e9
			}
		}
	}

	for k := 0; k < len(a); k++ {
		for i := 0; i < len(a); i++ {
			for j := 0; j < len(a[i]); j++ {
				if w[i][k] + w[k][j] < w[i][j] {
					w[i][j] = w[i][k] + w[k][j]
				}
			}
		}
	}

	return w
}
