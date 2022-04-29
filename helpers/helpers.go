package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"gitlab.com/toby3d/telegraph"
)

const (
	telegraphUploadAPI = "https://telegra.ph/upload"
)

func WriteFileJSON(filename string, obj interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error opening file")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(obj); err != nil {
		return fmt.Errorf("Error writing to file")
	}

	return nil
}

func CreateDomFromImages(images map[string]string) []telegraph.Node {
	result := make([]telegraph.Node, 0, len(images))
	for _, image := range images {
		result = append(result, telegraph.NodeElement{
			Tag:      "img",
			Attrs:    map[string]string{"src": image},
			Children: nil,
		})
	}

	return result
}

func PostImage(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return "", err
	}

	io.Copy(part, file)
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
	var resp []responseOK

	if err := json.Unmarshal(content, &resp); err != nil {
		var errResp responseErr

		if errErr := json.Unmarshal(content, &errResp); errErr != nil {
			return "", fmt.Errorf("%v - while trying to hanle - %v", errErr, err)
		}

		return "", fmt.Errorf("Error response: %v", errResp.Err)
	}

	if len(resp) != 1 {
		return "", fmt.Errorf("response has invalid length - %d", len(resp))
	}

	return resp[0].Src, nil
}

type responseOK struct {
	Src string `json:"src"`
}

type responseErr struct {
	Err string `json:"error"`
}
