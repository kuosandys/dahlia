package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kuosandys/dahlia/internal/configs"
	"github.com/kuosandys/dahlia/internal/dropbox"
	"github.com/kuosandys/dahlia/internal/generator"
)

const (
	CONFIG_FILENAME  = "config"
	CONFIG_FILEPATH  = "."
	DEFAULT_INTERVAL = 24 // hours
	OUTPUT_DIR       = "./dist/"
)

type application struct {
	configs       configs.Configs
	dropboxClient *dropbox.DropboxClient
	errorLog      *log.Logger
	infoLog       *log.Logger
}

func (a *application) run(upload bool) {
	g := generator.New(a.configs.URLs, a.configs.Interval)
	buf := new(bytes.Buffer)
	articleCount, fileName, err := g.GenerateEpub(buf)
	if err != nil {
		a.errorLog.Fatalf("Error generating file: %s", err)
	}

	if articleCount == 0 {
		a.infoLog.Println("Skipping generation: no new articles.")
		return
	}

	a.infoLog.Printf("Generated: %d new articles.", articleCount)

	// for testing
	if !upload {
		err := os.MkdirAll(OUTPUT_DIR, os.ModePerm)
		if err != nil {
			a.errorLog.Fatalf("Error creating %s directory: %s", OUTPUT_DIR, err)
		}

		path := fmt.Sprintf("%s%s", OUTPUT_DIR, fileName)
		err = os.WriteFile(path, buf.Bytes(), os.ModePerm)
		if err != nil {
			a.errorLog.Fatalf("Error writing file: %s", err)
		}

		a.infoLog.Printf("File written to %s", path)
		return
	}

	err = a.dropboxClient.GetAccessToken()
	if err != nil {
		a.errorLog.Fatalf("Error getting Dropbox access token: %s", err)
	}

	path, err := a.dropboxClient.Upload(a.configs.DropboxKoboFolder+fileName, buf)
	if err != nil {
		a.errorLog.Fatalf("Error uploading file to Dropbox: %s", err)
	}

	a.infoLog.Printf("File uploaded to %s", path)
}

func main() {
	dropboxAppKey := flag.String("dropboxAppKey", "", "Dropbox app key")
	dropboxAppSecret := flag.String("dropboxAppSecret", "", "Dropbox app secret")
	dropboxRefreshToken := flag.String("dropboxRefreshToken", "", "Dropbox refresh token")
	upload := flag.Bool("upload", true, "Whether to upload to Dropbox")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime|log.Lshortfile)

	configs, err := configs.Load(CONFIG_FILENAME, CONFIG_FILEPATH, DEFAULT_INTERVAL)
	if err != nil {
		errorLog.Fatal()
	}

	app := &application{
		configs:       configs,
		dropboxClient: dropbox.New(*dropboxAppKey, *dropboxAppSecret, *dropboxRefreshToken),
		errorLog:      errorLog,
		infoLog:       infoLog,
	}

	app.run(*upload)
}
