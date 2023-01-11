package telegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/bohdanch-w/go-tgupload/entities"
)

func (s *Server) Upload(ctx context.Context, media entities.MediaFile) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", media.Name)
	if err != nil {
		return "", fmt.Errorf("create multipart writer: %w", err)
	}

	if _, err := part.Write(media.Data); err != nil {
		return "", fmt.Errorf("encode multipart: %w", err)
	}

	writer.Close()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, telegraphUploadAPI, body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	return readResponse(content)
}

func readResponse(content []byte) (string, error) {
	var resp []struct {
		Src string `json:"src"`
	}

	if err := json.Unmarshal(content, &resp); err != nil {
		var errResp struct {
			Err string `json:"error"`
		}

		if errErr := json.Unmarshal(content, &errResp); errErr != nil {
			return "", fmt.Errorf("parse error response: %w", errErr)
		}

		return "", fmt.Errorf("error response: %w", entities.Error(errResp.Err))
	}

	if len(resp) != 1 || resp[0].Src == "" {
		return "", fmt.Errorf("invalid response")
	}

	return resp[0].Src, nil
}
