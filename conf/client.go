package conf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/zoenion/common/errors"
	"io/ioutil"
	"log"
	"net/http"
)

type Client struct {
	uri       string
	basicAuth string
}

func NewClient(uri, apiKey, apiSecret string) *Client {
	return &Client{
		uri:       uri,
		basicAuth: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", apiKey, apiSecret))),
	}
}

func (c *Client) Set(name string, conf Map) error {
	client := &http.Client{}
	confBytes, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	fullURI := fmt.Sprintf("%s%s%s", c.uri, RouteSet, name)
	req, err := http.NewRequest(http.MethodPost, fullURI, bytes.NewBuffer(confBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Basic "+c.basicAuth)
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode != http.StatusOK {
			log.Println("response status: ", resp.Status, fullURI)
			err = errors.New("failed to set configs")
		}
	}
	return err
}

func (c *Client) Get(name string) (Map, error) {
	var m Map
	client := &http.Client{}

	fullURI := fmt.Sprintf("%s%s%s", c.uri, RouteGet, name)
	req, err := http.NewRequest(http.MethodGet, fullURI, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+c.basicAuth)
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode != http.StatusOK {
			log.Println("response status: ", resp.Status, fullURI)
			err = errors.New("failed to set configs")
		} else {
			m = Map{}
			responseBytes, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				err = json.Unmarshal(responseBytes, &m)
			}
		}
	}
	return m, err
}

func (c *Client) Del(name string) error {
	client := &http.Client{}

	fullURI := fmt.Sprintf("%s%s%s", c.uri, RouteDel, name)
	req, err := http.NewRequest(http.MethodDelete, fullURI, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Basic "+c.basicAuth)
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode != http.StatusOK {
			log.Println("response status: ", resp.Status, fullURI)
			err = errors.New("failed to set configs")
		}
	}
	return err
}
