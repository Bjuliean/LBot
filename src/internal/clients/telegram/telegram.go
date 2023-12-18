package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

const(
	getUpdatesMethod = "getUpdates"
	sendMessageMethod = "sendMessage"
)

type Client struct {
	host	 string
	basePath string
	client	 http.Client
}

func New(host string, token string) *Client {
	return &Client{
		host: host,
		basePath: "bot" + token,
		client: http.Client{},
	}
}

func (c *Client)Updates(offset int, limit int) ([]Update, error) {
	const ferr = "clients.telegram.Updates"

	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(getUpdatesMethod, q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ferr, err)
	}

	var res UpdatesResponse

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("%s: %w", ferr, err)
	}
	return res.Result, nil
}

func (c *Client)SendMessage(chatID int, text string) error {
	const ferr = "clients.telegram.SendMessage"
	
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)

	_, err := c.doRequest(sendMessageMethod, q)
	if err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}
	return nil
}

func (c *Client)doRequest(method string, query url.Values) ([]byte, error) {
	const ferr = "clients.telegram.doRequest"
	
	u := url.URL{
		Scheme: "http",
		Host: c.host,
		Path: path.Join(c.basePath, method),
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ferr, err)
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ferr, err)
	}

	defer func(){_=resp.Body.Close()}()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ferr, err)
	}
	return body, nil
}