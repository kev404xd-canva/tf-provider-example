package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

func (c *APIClient) CreateTarget(target Target) (*Target, error) {
	rb, err := json.Marshal(target)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/targets", c.baseURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// Server will generate uuid for new target
	// Unmarshall response body for full target object
	createdTarget := Target{}
	err = json.Unmarshal(body, &createdTarget)
	if err != nil {
		return nil, err
	}

	return &createdTarget, nil
}

func (c *APIClient) GetTarget(id string) (*Target, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/target/%s", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	target := Target{}
	err = json.Unmarshal(body, &target)
	if err != nil {
		return nil, err
	}

	return &target, nil
}

func (c *APIClient) UpdateTarget(id string, target Target) (*Target, error) {
	rb, err := json.Marshal(target)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/targets/%s", c.baseURL, id), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	updatedTarget := Target{}
	err = json.Unmarshal(body, &updatedTarget)
	if err != nil {
		return nil, err
	}

	return &updatedTarget, nil
}

func (c *APIClient) DeleteTarget(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/targets/%s", c.baseURL, id), nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return err
	}

	expectedMessage := `{"message":"Target deleted successfully"}`
	if string(body) != expectedMessage {
		return errors.New(string(body))
	}

	return nil
}
