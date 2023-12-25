package send

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	"go.uber.org/zap"
)

// Client структура возвращаемая для работы, клиент
type Client struct {
	URL    string
	Client *http.Client
	Logger *zap.Logger
}

// New возвращает эксземпляр клиента
func New(url string, logger *zap.Logger) (*Client, error) {
	var client Client
	client.URL = url
	client.Client = &http.Client{
		Transport: &http.Transport{},
	}
	client.Logger = logger
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
	c.Logger.Info("JSON ready", zap.ByteString("json", data))

	req, err := http.NewRequest("POST", c.URL+"/update/", bytes.NewReader(data))
	req.Close = true // Close the connection after sending the request
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	b, err := io.ReadAll(resp.Body)
	c.Logger.Info("Response ready", zap.ByteString("response", b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}
