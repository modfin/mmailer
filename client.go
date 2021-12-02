package mmailer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	url        string
	httpClient *http.Client
}

func NewClient(url string, key string) *Client {
	return &Client{
		url:        url + "/send?key=" + key,
		httpClient: http.DefaultClient,
	}
}

func (c *Client) SetHttpClient(client *http.Client) {
	c.httpClient = client
}

func (c *Client) Send(ctx context.Context, e Email) (resps []Response, err error) {
	return c.SendWith(ctx, e, "")
}

func (c *Client) SendWith(ctx context.Context, e Email, service string) (resps []Response, err error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	if c.httpClient == nil {
		return nil, errors.New("no http client i available")
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json")
	if len(service) > 0 {
		req.Header.Set("X-Service", service)
	}

	res, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("did not get 200 from mmailer, %s", string(body))
	}
	err = json.Unmarshal(body, &resps)
	return resps, err
}
