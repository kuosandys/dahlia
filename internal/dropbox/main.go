package dropbox

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	ENDPOINT_TOKEN           = "https://api.dropbox.com/oauth2/token"
	ENDPOINT_FILES_UPLOAD    = "https://content.dropboxapi.com/2/files/upload"
	DEFAULT_FILE_UPLOAD_PATH = "/Apps/Rakuten Kobo/"
)

type DropboxClient struct {
	appKey       string
	appSecret    string
	refreshToken string
	accessToken  string
	httpClient   *http.Client
}

func New(appKey, appSecret, refreshToken string) *DropboxClient {
	return &DropboxClient{
		appKey:       appKey,
		appSecret:    appSecret,
		refreshToken: refreshToken,
		httpClient:   &http.Client{},
	}
}

// files must be smaller than 150 MB
func (d *DropboxClient) Upload(path string, r io.Reader) (string, error) {
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

	req, err := http.NewRequest(http.MethodPost, ENDPOINT_FILES_UPLOAD, r)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.accessToken))
	req.Header.Set("Dropbox-API-Arg", string(params))

	res, err := d.execute(req)
	if err != nil {
		return "", err
	}

	return res["path_display"].(string), nil
}

func (d *DropboxClient) GetAccessToken() error {
	body := url.Values{}
	body.Add("grant_type", "refresh_token")
	body.Add("client_id", d.appKey)
	body.Add("client_secret", d.appSecret)
	body.Add("refresh_token", d.refreshToken)

	req, err := http.NewRequest(http.MethodPost, ENDPOINT_TOKEN, strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}

	res, err := d.execute(req)
	if err != nil {
		return err
	}

	d.accessToken = res["access_token"].(string)

	return nil
}

func (d *DropboxClient) execute(req *http.Request) (map[string]interface{}, error) {
	res, err := d.httpClient.Do(req)
	if err != nil || res.StatusCode >= 300 {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var jsonRes map[string]interface{}
	err = json.Unmarshal(body, &jsonRes)
	if err != nil {
		return nil, err
	}

	return jsonRes, nil
}
