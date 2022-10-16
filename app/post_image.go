package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

const (
	telegraphUploadAPI = "https://telegra.ph/upload"
)

func postImage(name string, data []byte) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return "", err
	}

	part.Write(data)
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
