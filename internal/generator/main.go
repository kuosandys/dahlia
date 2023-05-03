package generator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/mmcdole/gofeed"
)

const (
	TEMPLATES_PATH = "./templates/"
	OUTPUT_PATH    = "./dist/"
	YYYYMMDD       = "2006-01-02"
)

type Generator struct {
	feedData map[string][]*gofeed.Item
}

type Data struct {
	FeedData      map[string][]*gofeed.Item
	DateFormatted string
}

func New() *Generator {
	return &Generator{
		feedData: make(map[string][]*gofeed.Item),
	}
}

func (g Generator) GenerateNewsletter(urls []string, lastHours int) (int, error) {
	err := g.getDataFromFeeds(urls, lastHours)
	if err != nil {
		return 0, err
	}

	err = g.templatePage()
	if err != nil {
		return 0, err
	}

	return g.getArticlesCount(), nil
}

func (g Generator) getDataFromFeeds(urls []string, lastHours int) error {
	fp := gofeed.NewParser()

	for _, url := range urls {
		feed, err := fp.ParseURL(url)
		if err != nil {
			return err
		}

		items := []*gofeed.Item{}

		for _, item := range feed.Items {
			if (time.Now()).Sub(*item.PublishedParsed).Hours() < float64(lastHours) {
				items = append(items, item)
			}
		}

		g.feedData[feed.Title] = items
	}

	return nil
}

func (g Generator) getArticlesCount() int {
	count := 0
	for _, v := range g.feedData {
		count += len(v)
	}
	return count
}

func (g Generator) templatePage() error {
	date := time.Now().UTC().Format(YYYYMMDD)
	filePath := filepath.Join(OUTPUT_PATH + date + ".html")
	f, err := os.Create(filePath)

	defer f.Close()

	if err != nil {
		return fmt.Errorf("Error creating file %s: %v", filePath, err)
	}

	w := bufio.NewWriter(f)
	t, err := g.loadTemplates()
	if err != nil {
		return err
	}

	data := Data{
		FeedData:      g.feedData,
		DateFormatted: date,
	}

	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("Error executing template %s : %v", filePath, err)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("Error writing file %s: %v", filePath, err)
	}

	return nil
}

func (g Generator) loadTemplates() (*template.Template, error) {
	files := []string{"base.tmpl", "content.tmpl"}

	var paths []string
	for _, tmpl := range files {
		paths = append(paths, filepath.Join(TEMPLATES_PATH, tmpl))
	}

	t, err := template.ParseFiles(paths...)
	if err != nil {
		return nil, err
	}

	return t, nil
}
