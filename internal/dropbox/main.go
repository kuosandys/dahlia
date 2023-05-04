package dropbox

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	DROPBOX_API              = "https://content.dropboxapi.com/2"
	ENDPOINT_FILES_UPLOAD    = "/files/upload"
	DEFAULT_FILE_UPLOAD_PATH = "/Apps/Rakuten Kobo/"
)

type DropboxFilesClient struct {
	HttpClient  *http.Client
	AccessToken string
}

func New(accessToken string) *DropboxFilesClient {
	return &DropboxFilesClient{
		HttpClient:  &http.Client{},
		AccessToken: accessToken,
	}
}

// files must be smaller than 150 MB
func (d *DropboxFilesClient) Upload(path string, r io.Reader) (string, error) {
	params, err := json.Marshal(map[string]interface{}{
		"autorename":      false,
		"mode":            "add",
		"mute":            false,
		"path":            path,
		"strict_conflict": false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, DROPBOX_API+ENDPOINT_FILES_UPLOAD, r)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.AccessToken))
	req.Header.Set("Dropbox-API-Arg", string(params))

	res, err := d.execute(req)
	if err != nil {
		return "", err
	}

	var jsonRes map[string]interface{}
	err = json.Unmarshal(res, &jsonRes)
	if err != nil {
		return "", err
	}
	return jsonRes["path_display"].(string), nil
}

func (d *DropboxFilesClient) execute(req *http.Request) ([]byte, error) {
	res, err := d.HttpClient.Do(req)
	if err != nil || res.StatusCode >= 300 {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
