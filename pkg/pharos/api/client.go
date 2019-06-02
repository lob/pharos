package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/lob/pharos/pkg/pharos/config"
	"github.com/pkg/errors"
)

// Client is a struct containing information for an api client.
type Client struct {
	client *http.Client
	config *config.Config
}

// NewClient creates a new Client with its own http.Client.
func NewClient(config *config.Config) *Client {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &Client{c, config}
}

// ClientFromConfig creates a new Client with its own http.Client
// using the config file provided.
func ClientFromConfig(configFile string) (*Client, error) {
	c, err := config.New(configFile)
	if err != nil {
		return nil, err
	}
	err = c.Load()
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// send sends a http.Request for the specified method and path, with the given body encoded as JSON.
// It then marshalls the returned response into the given response interface.
func (c *Client) send(method string, path string, body interface{}, response interface{}, query map[string]string) error {
	buf := &bytes.Buffer{}
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return err
		}
	}

	// Create http request with json body.
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.config.BaseURL, path), buf)
	if err != nil {
		return errors.Wrap(err, "unable to create http request")
	}
	req.Header.Set("Content-Type", "application/json")

	// Add queries to request if there are any.
	if query != nil {
		q := url.Values{}
		for key, value := range query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Send request.
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send http request")
	}
	defer resp.Body.Close()

	err = checkError(resp)
	if err != nil {
		return errors.Wrap(err, "response contained error")
	}

	// Parse response body.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read response body")
	}

	if err = json.Unmarshal(respBody, response); err != nil {
		return errors.Wrap(err, "could not unmarshal response into interface")
	}

	return nil
}

func checkError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	errMsg := new(struct {
		Err struct {
			Message string `json:"message"`
		} `json:"error"`
	})

	err := json.NewDecoder(resp.Body).Decode(errMsg)
	if err != nil {
		return errors.Wrap(err, http.StatusText((resp.StatusCode)))
	}

	return fmt.Errorf("%s (%d)", errMsg.Err.Message, resp.StatusCode)
}
