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
	DATE_FORMAT    = "2006 Jan 2"
	FILE_NAME      = "index"
	OUTPUT_PATH    = "./dist/"
	TEMPLATES_PATH = "./templates/"
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

	err = g.templatePage(lastHours)
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
				// try format date
				if published, err := time.Parse(time.RFC1123, item.Published); err == nil {
					item.Published = published.Format(DATE_FORMAT)
				}
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

func (g Generator) templatePage(lastHours int) error {
	dateString := fmt.Sprintf("%s - %s", time.Now().Add(-time.Hour*time.Duration(lastHours)).UTC().Format(DATE_FORMAT), time.Now().UTC().Format(DATE_FORMAT))
	filePath := filepath.Join(OUTPUT_PATH + FILE_NAME + ".html")
	err := os.MkdirAll(OUTPUT_PATH, os.ModePerm)
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
		DateFormatted: dateString,
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
