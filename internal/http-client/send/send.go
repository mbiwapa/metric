package send

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
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
	const op = "http-client.send.Send"
	c.Logger.With(zap.String("op", op))

	body := format.Metrics{
		MType: typ,
		ID:    name,
	}

	switch typ {
	case format.Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			c.Logger.Error("Cant parse gauge metric", zap.Error(err))
			return err
		}
		body.Value = &val
	case format.Counter:
		val, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			c.Logger.Error("Cant parse counter metric", zap.Error(err))
			return err
		}
		body.Delta = &val
	default:
	}

	data, err := json.Marshal(body)
	if err != nil {
		c.Logger.Error("Cant encoding request", zap.Error(err))
		return err
	}
	c.Logger.Info("JSON ready", zap.ByteString("json", data))

	req, err := http.NewRequest("POST", c.URL+"/update/", bytes.NewReader(data))
	if err != nil {
		c.Logger.Error("Cant create request", zap.Error(err))
		return err
	}
	req.Close = true // Close the connection after sending the request

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	//FIXME именно в автотестах клиент периодически ловит EOF и отваливается
	//FIXME как правило на первый же запрос, но потом приходит в себя, (Уточнить у ментора)
	//TODO судя по всему сервер просто долго поднимается.... убил на это больше 5 часов
	if err != nil {
		c.Logger.Error("Cant send metric", zap.Error(err))
	}

	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			c.Logger.Error("Cant send metric", zap.String("error", resp.Status))
		}
	}
	return nil
}
