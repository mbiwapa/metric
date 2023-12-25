package send

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/mbiwapa/metric/internal/lib/api/format"
)

// Client структура возвращаемая для работы, клиент
type Client struct {
	URL    string
	Client http.Client
}

// New возвращает эксземпляр клиента
func New(url string) (*Client, error) {
	var client Client
	client.URL = url
	client.Client = http.Client{}
	return &client, nil
}

// Send отправляет метрику на сервер
func (c *Client) Send(typ string, name string, value string) error {

	body := format.Metrics{
		MType: typ,
		ID:    name,
	}

	switch typ {
	case format.Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		body.Value = &val
	case format.Counter:
		val, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			return err
		}
		body.Delta = &val
	default:
	}

	data, err := json.Marshal(body)

	resp, err := c.Client.Post(
		c.URL+"/update/",
		"application/json",
		bytes.NewReader(data),
	)
	if resp != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}
