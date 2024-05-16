package client

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/agent/compressor"
	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/lib/retry/backoff"
	"github.com/mbiwapa/metric/internal/lib/signature"
)

// Client структура возвращаемая для работы, клиент
type Client struct {
	URL        string
	Client     *http.Client
	Logger     *zap.Logger
	Compressor *compressor.Compressor
	Key        string //ключ для вычисления хеша sha256
}

// New возвращает эксземпляр клиента
func New(url string, key string, logger *zap.Logger) (*Client, error) {
	var client Client
	client.URL = url
	client.Client = &http.Client{
		Transport: &http.Transport{},
	}
	client.Logger = logger
	client.Compressor = compressor.New(logger)
	client.Key = key

	return &client, nil
}

// Send отправляет метрику на сервер
func (c *Client) Send(gauges [][]string, counters [][]string) error {
	const op = "http-client.send.Send"
	logger := c.Logger.With(zap.String("op", op))

	var body []format.Metric

	for _, gauge := range gauges {
		val, err := strconv.ParseFloat(gauge[1], 64)
		if err != nil {
			logger.Error("Cant parse gauge metric", zap.Error(err))
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
			logger.Error("Cant parse counter metric", zap.Error(err))
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
		logger.Error("Cant encoding request", zap.Error(err))
		return err
	}
	compressedData, err := c.Compressor.GetCompressedData(data)
	if err != nil {
		logger.Error("Cant initializing compressed reader", zap.Error(err))
		return err
	}

	action := func(attempt uint) error {
		req, err := http.NewRequest("POST", c.URL+"/updates/", compressedData)
		if err != nil {
			logger.Error("Cant create request", zap.Error(err), zap.Uint("attempt", attempt))
			return err
		}
		req.Close = true // Close the connection after sending the request

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		if c.Key != "" {
			hashStr := signature.GetHash(c.Key, string(data), logger)
			logger.Info("Hash is generated", zap.String("hash", hashStr))
			req.Header.Set("HashSHA256", hashStr)
		}

		resp, err := c.Client.Do(req)
		if err != nil {
			logger.Error("Cant send metric", zap.Error(err), zap.Uint("attempt", attempt))
			return err
		}

		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				logger.Error("No response", zap.String("error", resp.Status), zap.Uint("attempt", attempt))
			}
			return err
		}
		logger.Info("Request completed successfully!", zap.ByteString("json", data))
		return nil
	}

	err = retry.Retry(
		action,
		strategy.Limit(4),
		strategy.Backoff(backoff.Backoff()))

	if err != nil {
		logger.Error("Cant send metric affter 4 attemt", zap.Error(err))
	}

	return nil
}

// Worker отправляет метрику на сервер в режиме потока
func (c *Client) Worker(jobs <-chan map[string][][]string, errorChanel chan<- error) {
	for j := range jobs {
		err := c.Send(j["gauge"], j["counter"])
		if err != nil {
			errorChanel <- err
		}
	}
}
