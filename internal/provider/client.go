package provider

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type APIClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

func NewAPIClient(host, username, password string) *APIClient {
	return &APIClient{
		baseURL:  host,
		username: username,
		password: password,
		client:   &http.Client{},
	}
}

func (c *APIClient) doRequest(req *http.Request) ([]byte, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
