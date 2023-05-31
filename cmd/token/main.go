package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/kuosandys/anthology/internal/dropbox"
	"github.com/pkg/browser"
)

const (
	ADDR          = "localhost:5000"
	REDIRECT_PATH = "/redirect"
	REDIRECT_URI  = "http://localhost:5000/redirect"
)

type application struct {
	dropboxClient *dropbox.DropboxClient
	errorLog      *log.Logger
	infoLog       *log.Logger
}

func (a *application) run(dropboxAppKey string) {
	params := url.Values{}
	params.Add("client_id", dropboxAppKey)
	params.Add("token_access_type", "offline")
	params.Add("response_type", "code")
	params.Add("redirect_uri", REDIRECT_URI)

	url := url.URL{
		Scheme:   "https",
		Host:     "www.dropbox.com",
		Path:     "/oauth2/authorize",
		RawQuery: params.Encode(),
	}
	browser.OpenURL(url.String())
}

func main() {
	dropboxAppKey := flag.String("dropboxAppKey", "", "Dropbox app key")
	dropboxAppSecret := flag.String("dropboxAppSecret", "", "Dropbox app secret")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime|log.Lshortfile)

	app := &application{
		dropboxClient: dropbox.New(*dropboxAppKey, *dropboxAppSecret, ""),
		errorLog:      errorLog,
		infoLog:       infoLog,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(REDIRECT_PATH, app.handleRedirect)

	go func() {
		for {
			time.Sleep(time.Second)

			res, err := http.Get(fmt.Sprintf("http://%s", ADDR))
			if err != nil {
				log.Println("Cannot reach server:", err)
				continue
			}

			defer res.Body.Close()

			break
		}
		app.run(*dropboxAppKey)
	}()

	err := http.ListenAndServe(ADDR, mux)
	if err != nil {
		log.Fatal(err)
	}
}
