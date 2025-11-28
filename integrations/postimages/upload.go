package postimages

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/bohdanch-w/go-tgupload/entities"
)

const (
	uploadURL = "https://api.postimage.org/1/upload"

	o        = "2b819584285c102318568238c7d4a4c7"
	m        = "59c2ad4b46b0c1e12d5703302bff0120"
	version  = "1.0.1"
	portable = "1"
)

func NewAPI(apiKey string, gallery string) *API {
	return &API{
		apiKey: apiKey,
		cli:    http.DefaultClient,
	}
}

type API struct {
	cli     *http.Client
	apiKey  string
	gallery string
}

func (s *API) Upload(ctx context.Context, media entities.MediaFile) (string, error) {
	var (
		name      = filepath.Base(media.Name)
		idx       = strings.LastIndexByte(name, '.')
		nameNoExt = name
		ext       string

		form = make(url.Values)
	)

	if idx > 0 {
		nameNoExt, ext = name[:idx], name[idx+1:]
	}

	image := base64.StdEncoding.EncodeToString(media.Data)

	form.Add("o", o)
	form.Add("m", m)
	form.Add("name", nameNoExt)
	form.Add("type", ext)
	form.Add("version", version)
	form.Add("portable", portable)
	form.Add("key", s.apiKey)
	form.Add("image", image)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	return ParseResponse(content)
}

func ParseResponse(data []byte) (string, error) {
	var resp response

	if err := xml.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return resp.Links.Hotlink, nil
}

type response struct {
	XMLName xml.Name `xml:"data"`
	Success string   `xml:"success,attr"`
	Status  string   `xml:"status,attr"`
	Links   struct {
		Hotlink string `xml:"hotlink"`
	} `xml:"links"`
}
