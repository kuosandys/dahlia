package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/kuosandys/dahlia/internal/configs"
	"github.com/kuosandys/dahlia/internal/dropbox"
	"github.com/kuosandys/dahlia/internal/generator"
)

const (
	CONFIG_FILENAME  = "config"
	CONFIG_FILEPATH  = "."
	DEFAULT_INTERVAL = 24 // hours
)

type application struct {
	configs  configs.Configs
	errorLog *log.Logger
	infoLog  *log.Logger
}

func (a *application) run(dropboxAccessToken string) {
	g := generator.New()
	articles, err := g.GenerateNewsletter(a.configs.URLs, a.configs.Interval)
	if err != nil {
		a.errorLog.Fatal(err)
	}

	if articles > 0 {
		a.infoLog.Printf("Newsletter generated: %d new articles.", articles)
	} else {
		a.infoLog.Println("Skipping newsletter generation: no new articles.")
	}

	client := dropbox.New(dropboxAccessToken)
	path, err := client.Upload(a.configs.DropboxKoboFolder+"test.txt", strings.NewReader("test"))
	if err != nil {
		a.errorLog.Fatal(err)
	}
	a.infoLog.Printf("File saved to %s", path)
}

func main() {
	dropboxAccessToken := flag.String("dropboxAccessToken", "", "Dropbox access token")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime|log.Lshortfile)

	configs, err := configs.Load(CONFIG_FILENAME, CONFIG_FILEPATH, DEFAULT_INTERVAL)
	if err != nil {
		errorLog.Fatal()
	}

	app := &application{
		infoLog:  infoLog,
		errorLog: errorLog,
		configs:  configs,
	}

	app.run(*dropboxAccessToken)
}
