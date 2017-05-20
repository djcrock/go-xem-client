package xem

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Available origin types
const (
	AniDB = "anidb"
	Scene = "scene"
	TVDB  = "tvdb"
)

// XEM API URL strings and response constants
const (
	defaultUserAgent     = "go-xem-client/0.1"
	defaultBaseURL       = "http://thexem.de/"
	defaultAllEndpoint   = "map/all"
	defaultNamesEndpoint = "map/allNames"

	// Response success indicator
	success = "success"
)

// Mapping from origins to Episodes
type Mapping map[string]Episode

// Episode numbering, including both season and absolute episode numbers.
type Episode struct {
	Season   int `json:"season"`
	Episode  int `json:"episode"`
	Absolute int `json:"absolute"`
}

// Client for the XEM API
type Client struct {
	client *http.Client

	UserAgent     string
	BaseURL       *url.URL
	AllEndpoint   *url.URL
	NamesEndpoint *url.URL
}

// NewClient creates a new XEM API client
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(defaultBaseURL)
	allEndpoint, _ := url.Parse(defaultAllEndpoint)
	namesEndpoint, _ := url.Parse(defaultNamesEndpoint)

	c := &Client{
		client:        httpClient,
		BaseURL:       baseURL,
		AllEndpoint:   allEndpoint,
		NamesEndpoint: namesEndpoint,
	}

	return c
}

// NewRequest creats an API request.
func (c *Client) NewRequest(method string, resURL *url.URL) (*http.Request, error) {
	u := c.BaseURL.ResolveReference(resURL)

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}

	return req, nil
}

type allResponse struct {
	Result  string    `json:"result"`
	Data    []Mapping `json:"data"`
	Message string    `json:"message"`
}

// All retrieves all mappings from the given origin and ID
func (c *Client) All(origin, id string) ([]Mapping, error) {
	vals := make(url.Values)
	vals.Set("origin", origin)
	vals.Set("id", id)
	c.AllEndpoint.RawQuery = vals.Encode()

	all := &allResponse{}
	_, err := c.get(c.AllEndpoint, all)
	if err != nil {
		return nil, err
	}
	if all.Result != success {
		return nil, fmt.Errorf("request failed: %v", all.Message)
	}

	return all.Data, nil
}

type namesResponse struct {
	Result  string                        `json:"result"`
	Data    map[string]([]map[string]int) `json:"data"`
	Message string                        `json:"message"`
}

// Names retrieves the names of
func (c *Client) Names(origin, lang string) (map[string]([]map[string]int), error) {
	vals := make(url.Values)
	vals.Set("origin", origin)
	vals.Set("seasonNumbers", "1")
	vals.Set("language", lang)
	c.NamesEndpoint.RawQuery = vals.Encode()

	all := &namesResponse{}
	_, err := c.get(c.NamesEndpoint, all)
	if err != nil {
		return nil, err
	}
	if all.Result != success {
		return nil, fmt.Errorf("request failed: %v", all.Message)
	}

	return all.Data, nil
}

func (c *Client) get(endpoint *url.URL, result interface{}) (*http.Response, error) {
	req, err := c.NewRequest("GET", endpoint)
	if err != nil {
		return nil, err
	}
	r, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return r, fmt.Errorf("unable to read response body: %v", err)
	}

	if r.StatusCode < 200 || r.StatusCode > 299 {
		return r, fmt.Errorf("%v: %d %s", r.Request.URL, r.StatusCode, string(data))
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		return r, fmt.Errorf("unable to decode JSON: %v %s", err, string(data))
	}

	return r, nil
}
