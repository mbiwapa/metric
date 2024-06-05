// Package client provides a client that sends metrics to a server.
// It contains the necessary configurations such as URL, HTTP client, logger, compressor, and a key for hash generation.
// It also provides functions to send metrics to the server and to run a worker that continuously reads jobs from a channel and sends the metrics using the Send method.
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

// Client represents a client that sends metrics to a server.
// It contains the necessary configurations such as URL, HTTP client, logger, compressor, and a key for hash generation.
type Client struct {
	URL        string                 // URL is the base URL for the client to send requests to.
	Client     *http.Client           // Client is the HTTP client used to send requests.
	Logger     *zap.Logger            // Logger is used for logging purposes.
	Compressor *compressor.Compressor // Compressor is used to compress the data before sending.
	Key        string                 // Key is used for generating SHA256 hashes for request validation.
}

// New initializes and returns a new instance of the Client struct.
// It sets up the URL, HTTP client, logger, compressor, and key for the client.
//
// Parameters:
//   - url: The base URL for the client to send requests to.
//   - key: The key used for generating SHA256 hashes for request validation.
//   - logger: A zap.Logger instance for logging purposes.
//
// Returns:
//   - *Client: A pointer to the newly created Client instance.
//   - error: An error if there is an issue during the creation of the Client instance.
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

// Send sends metrics to the server. It takes gauge and counter metrics, processes them, compresses the data, and sends it to the server with retry logic.
//
// Parameters:
//   - gauges: A slice of slices containing gauge metrics, where each inner slice contains the metric ID and value as strings.
//   - counters: A slice of slices containing counter metrics, where each inner slice contains the metric ID and value as strings.
//
// Returns:
//   - error: An error if there is an issue during the processing or sending of the metrics.
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

	data, errJSON := json.Marshal(body)
	if errJSON != nil {
		logger.Error("Cant encoding request", zap.Error(errJSON))
		return errJSON
	}
	compressedData, errCompress := c.Compressor.GetCompressedData(data)
	if errCompress != nil {
		logger.Error("Cant initializing compressed reader", zap.Error(errCompress))
		return errCompress
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

	err := retry.Retry(
		action,
		strategy.Limit(4),
		strategy.Backoff(backoff.Backoff()))

	if err != nil {
		logger.Error("Cant send metric affter 4 attemt", zap.Error(err))
	}

	return nil
}

// Worker sends metrics to the server in a streaming mode. It continuously reads jobs from the provided channel and sends the metrics using the Send method.
//
// Parameters:
//   - jobs: A channel that provides jobs, where each job is a map containing gauge and counter metrics.
//   - errorChanel: A channel to send errors if there is an issue during the processing or sending of the metrics.
func (c *Client) Worker(jobs <-chan map[string][][]string, errorChanel chan<- error) {
	for j := range jobs {
		err := c.Send(j["gauge"], j["counter"])
		if err != nil {
			errorChanel <- err
		}
	}
}
