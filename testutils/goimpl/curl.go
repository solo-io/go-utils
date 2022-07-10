package goimpl

import (
	"bytes"
	"io"
	"net/http"
)

// Curl provides some of the functionality you would expect from the curl command.
// This saves you from having to call the curl binary in environments where that binary is not available.
func Curl(url string) (string, error) {
	body := bytes.NewReader([]byte(url))
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		return "", err
	}

	return ExecuteRequest(req)
}

// ExecuteRequest provides some of the functionality you would expect from the curl command.
// This saves you from having to call the curl binary in environments where that binary is not available.
func ExecuteRequest(request *http.Request) (string, error) {
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	p := new(bytes.Buffer)
	_, err = io.Copy(p, resp.Body)
	defer resp.Body.Close()

	return p.String(), nil
}
