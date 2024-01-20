package client

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/agent/compressor"
	"github.com/mbiwapa/metric/internal/lib/api/format"
)

// Client структура возвращаемая для работы, клиент
type Client struct {
	URL        string
	Client     *http.Client
	Logger     *zap.Logger
	Compressor *compressor.Compressor
}

// New возвращает эксземпляр клиента
func New(url string, logger *zap.Logger) (*Client, error) {
	var client Client
	client.URL = url
	client.Client = &http.Client{
		Transport: &http.Transport{},
	}
	client.Logger = logger
	client.Compressor = compressor.New(logger)

	return &client, nil
}

// Send отправляет метрику на сервер
func (c *Client) Send(gauges [][]string, counters [][]string) error {
	const op = "http-client.send.Send"
	c.Logger.With(zap.String("op", op))

	var body []format.Metric

	for _, gauge := range gauges {
		val, err := strconv.ParseFloat(gauge[1], 64)
		if err != nil {
			c.Logger.Error("Cant parse gauge metric", zap.Error(err))
			return err
		}
		body = append(body, format.Metric{
			MType: format.Gauge,
			ID:    gauge[0],
			Value: &val,
		})
	}

	for _, counter := range counters {
		val, err := strconv.ParseInt(counter[1], 10, 64)
		if err != nil {
			c.Logger.Error("Cant parse counter metric", zap.Error(err))
			return err
		}
		body = append(body, format.Metric{
			MType: format.Counter,
			ID:    counter[0],
			Delta: &val,
		})
	}

	data, err := json.Marshal(body)
	if err != nil {
		c.Logger.Error("Cant encoding request", zap.Error(err))
		return err
	}
	compressedData, err := c.Compressor.GetCompressedData(data)
	if err != nil {
		c.Logger.Error("Cant initializing compressed reader", zap.Error(err))
		return err
	}
	c.Logger.Info("JSON ready", zap.ByteString("json", data))

	req, err := http.NewRequest("POST", c.URL+"/updates/", compressedData)
	if err != nil {
		c.Logger.Error("Cant create request", zap.Error(err))
		return err
	}
	req.Close = true // Close the connection after sending the request

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := c.Client.Do(req)
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
