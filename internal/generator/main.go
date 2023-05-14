package generator

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bmaupin/go-epub"
	"github.com/mmcdole/gofeed"
	"github.com/pgaskin/kepubify/kepub"
)

const (
	DATE_FORMAT = "2006 Jan 2"
)

type Generator struct {
	epub       *epub.Epub
	feedParser *gofeed.Parser
	feeds      []string
	images     map[string]string
	lastHours  int
}

type Article struct {
	title   string
	link    string
	author  string
	date    string
	content string
}

func New(feeds []string, lastHours int) *Generator {
	g := &Generator{
		feeds:      feeds,
		feedParser: gofeed.NewParser(),
		images:     make(map[string]string),
		lastHours:  lastHours,
	}
	g.epub = epub.NewEpub(g.getTitle())

	return g
}

func (g *Generator) GenerateKepub(buf *bytes.Buffer) (int, string, error) {
	tempEpub, err := ioutil.TempFile("", g.getTitle()+".epub")
	if err != nil {
		return 0, "", err
	}

	tempKepub, err := ioutil.TempFile("", g.getTitle()+"kepub.epub")
	if err != nil {
		return 0, "", err
	}

	defer os.Remove(tempEpub.Name())
	defer os.Remove(tempKepub.Name())

	articleCount, err := g.generate()
	if err != nil {
		return articleCount, "", err
	}

	// write to temp epub file - required by kepub converter
	err = g.epub.Write(tempEpub.Name())
	if err != nil {
		return articleCount, "", err
	}

	converter := kepub.Converter{}
	err = converter.Convert(tempEpub.Name(), tempKepub.Name())
	if err != nil {
		return articleCount, "", err
	}

	data, err := os.ReadFile(tempKepub.Name())
	if err != nil {
		return articleCount, "", err
	}

	_, err = buf.Write(data)
	if err != nil {
		return articleCount, "", err
	}

	return articleCount, g.getTitle() + ".kepub.epub", nil
}

func (g *Generator) GenerateEpub(buf *bytes.Buffer) (int, string, error) {
	articleCount, err := g.generate()
	if err != nil {
		return articleCount, "", err
	}

	_, err = g.epub.WriteTo(buf)
	if err != nil {
		return articleCount, "", err
	}

	return articleCount, g.getTitle() + ".epub", nil
}

func (g *Generator) generate() (int, error) {
	articleCount := 0
	articlesByFeed := make(map[string][]Article)

	wg := sync.WaitGroup{}
	wg.Add(len(g.feeds))

	for _, feed := range g.feeds {
		go func(feed string) {
			defer wg.Done()

			feedTitle, articles, err := g.getArticlesFromFeed(feed)
			if err != nil || len(articles) == 0 {
				return
			}

			articleCount += len(articles)
			articlesByFeed[feedTitle] = articles
		}(feed)
	}

	wg.Wait()

	sourcesContent := `
	<h2>Sources</h2>
	`
	for feedTitle, articles := range articlesByFeed {
		for _, article := range articles {
			sectionTitle := fmt.Sprintf(`
			<h2>%s</h2>
			<p>%s // %s // <a href="%s">Source</a></p>
			<hr></hr>
			`, article.title, feedTitle, article.date, article.link)

			content := article.content
			if document, err := goquery.NewDocumentFromReader(strings.NewReader(article.content)); err == nil {
				if url, err := url.Parse(article.link); err == nil {
					document = g.fixImages(document, fmt.Sprintf("%s://%s", url.Scheme, url.Host))
					content, _ = document.Html()
				}
			}

			g.epub.AddSection(html.UnescapeString(sectionTitle+content), article.title, "", "")

			sourcesContent += fmt.Sprintf(`
			<p>%s %s, <i>%s</i>, accessed %s, %s</p>
			`, article.author, article.date, feedTitle, time.Now().UTC().Format(DATE_FORMAT), article.link)
		}
	}

	if articleCount == 0 {
		return 0, nil
	}

	// cite sources
	g.epub.AddSection(html.UnescapeString(sourcesContent), "Sources", "", "")

	return articleCount, nil
}

func (g *Generator) getTitle() string {
	return fmt.Sprintf("%s - %s", time.Now().Add(-time.Hour*time.Duration(g.lastHours)).UTC().Format(DATE_FORMAT), time.Now().UTC().Format(DATE_FORMAT))
}

func (g *Generator) getArticlesFromFeed(url string) (string, []Article, error) {
	feed, err := g.feedParser.ParseURL(url)
	if err != nil {
		return "", nil, err
	}

	var author string
	if len(feed.Authors) > 0 {
		author = feed.Authors[0].Name
	}

	articles := []Article{}

	for _, item := range feed.Items {
		// assumption: feed is sorted from newest to oldest
		if time.Since(*item.PublishedParsed).Hours() > float64(g.lastHours) {
			break
		}

		// try get author if missing from feed
		if len(item.Authors) > 0 {
			author = item.Authors[0].Name
		}

		// try format date
		if published, err := time.Parse(time.RFC1123, item.Published); err == nil {
			item.Published = published.Format(DATE_FORMAT)
		}

		article := Article{
			title:   item.Title,
			link:    item.Link,
			author:  author,
			date:    item.Published,
			content: item.Content,
		}
		articles = append(articles, article)
	}

	return feed.Title, articles, nil
}

func (g *Generator) fixImages(document *goquery.Document, baseURL string) *goquery.Document {
	var err error

	document.Find("img").Each(func(i int, img *goquery.Selection) {
		src, exists := img.Attr("src")
		if !exists {
			return
		}

		relativePath, ok := g.images[src]
		if !ok {
			var imageURL = src
			// if src is a relative path, prefix with base URL
			if strings.HasPrefix(src, "/") {
				imageURL = baseURL + src
			}
			relativePath, err = g.epub.AddImage(imageURL, "")
			if err != nil {
				return
			}

			g.images[src] = relativePath
		}

		img.SetAttr("src", relativePath)
		img.RemoveAttr("srcset")
		img.RemoveAttr("loading")
	})

	return document
}
