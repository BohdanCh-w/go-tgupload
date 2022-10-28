package telegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/zumorl/go-tgupload/entities"
)

func (s *Server) Upload(ctx context.Context, media entities.MediaFile) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", media.Name)
	if err != nil {
		return "", err
	}

	part.Write(media.Data)
	writer.Close()

	request, err := http.NewRequest(http.MethodPost, telegraphUploadAPI, body)
	if err != nil {
		return "", err
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return readResponse(content)
}

func readResponse(content []byte) (string, error) {
	var resp []struct {
		Src string `json:"src"`
	}

	if err := json.Unmarshal(content, &resp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(resp) != 1 || resp[0].Src == "" {
		var errResp struct {
			Err string `json:"error"`
		}

		if errErr := json.Unmarshal(content, &errResp); errErr != nil {
			return "", fmt.Errorf("parse error response: %w", errErr)
		}

		return "", fmt.Errorf("Error response: %s", errResp.Err)
	}

	return resp[0].Src, nil
}
