package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kuosandys/anthology/internal/configs"
	"github.com/kuosandys/anthology/internal/dropbox"
	"github.com/kuosandys/anthology/internal/generator"
)

const (
	CONFIG_FILENAME    = "config"
	CONFIG_FILEPATH    = "."
	DEFAULT_LAST_HOURS = 168
	OUTPUT_DIR         = "./dist/"
)

type application struct {
	configs       configs.Configs
	dropboxClient *dropbox.DropboxClient
	errorLog      *log.Logger
	infoLog       *log.Logger
}

func (a *application) run(upload bool) {
	g := generator.New(a.configs.URLs, a.configs.LastHours)
	buf := new(bytes.Buffer)
	articleCount, fileName, err := g.GenerateKepub(buf)
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
	upload := flag.Bool("upload", false, "Whether to upload to Dropbox; defaults to 'false'")
	configFileName := flag.String("configFileName", CONFIG_FILENAME, "Config file name; defaults to 'config.yml'")
	configFilePath := flag.String("configFilePath", CONFIG_FILEPATH, "Path to config file; defaults to '.'")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime|log.Lshortfile)

	configs, err := configs.Load(*configFileName, *configFilePath, DEFAULT_LAST_HOURS)
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
