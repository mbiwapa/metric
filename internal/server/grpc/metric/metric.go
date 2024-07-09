package metric

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	config "github.com/mbiwapa/metric/internal/config/server"
	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/lib/signature"
	"github.com/mbiwapa/metric/internal/server/backuper"
	"github.com/mbiwapa/metric/internal/storage"
	pb "github.com/mbiwapa/metric/proto"
)

type MetricsServer struct {
	pb.UnimplementedMetricServiceServer
	conf    *config.Config
	logger  *zap.Logger
	storage storage.Storage
	backup  *backuper.Buckuper
}

func NewMetricServer(conf *config.Config, logger *zap.Logger, storage storage.Storage, backup *backuper.Buckuper) *MetricsServer {
	return &MetricsServer{
		conf:    conf,
		logger:  logger,
		storage: storage,
		backup:  backup,
	}
}

// Ping ping the DB
func (s *MetricsServer) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.storage.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.PingResponse{}, nil
}

// UpdateMetric update metric
func (s *MetricsServer) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	const op = "grpc.metric.UpdateMetric"

	response := pb.UpdateMetricResponse{}
	response.Metric = req.Metric

	log := s.logger.With(zap.String("op", op))

	// Check if the metric ID is empty
	if req.Metric.Id == "" {
		log.Error("Name is empty!")
		return nil, status.Errorf(codes.NotFound, `Metric %s not found`, req.Metric.Id)
	}

	// Create a context with a timeout for database operations
	databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var updateErr error

	// Update the metric based on its type
	switch req.Metric.Type {
	case format.Gauge:
		updateErr = s.storage.UpdateGauge(databaseCtx, req.Metric.Id, req.Metric.Value)
	case format.Counter:
		updateErr = s.storage.UpdateCounter(databaseCtx, req.Metric.Id, req.Metric.Delta)
		databaseGetCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		stringVal, err := s.storage.GetMetric(databaseGetCtx, req.Metric.Type, req.Metric.Id)
		if err != nil {
			log.Error("Failed to get metric", zap.Error(err))
			return nil, status.Errorf(codes.Internal, `Failed to get metric: %s`, req.Metric.Id)
		}
		newVal, err := strconv.ParseInt(stringVal, 0, 64)
		if err != nil {
			log.Error("Failed to parse int", zap.Error(err))
			return nil, status.Errorf(codes.Internal, `Failed to parse int: %s`, stringVal)
		}
		response.Metric.Delta = newVal
	default:
		log.Error("Undefined metric type", zap.String("type", req.Metric.Type))
		return nil, status.Errorf(codes.InvalidArgument, `Undefined metric type: %s`, req.Metric.Type)
	}

	// Handle update errors
	if updateErr != nil {
		log.Error("Failed to update value", zap.Error(updateErr))
		return nil, status.Errorf(codes.Internal, `Undefined metric type: %s`, req.Metric.Type)
	}

	//json marshal response.Metric
	marshal, err := json.Marshal(response.Metric)
	if err != nil {
		log.Error("Error encoding response", zap.Error(err))
		return nil, status.Error(codes.Internal, `Error encoding response`)
	}

	// Generate and set SHA256 hash if key is provided
	if s.conf.Key != "" {
		hashStr := signature.GetHash(s.conf.Key, string(marshal), log)
		_ = grpc.SetHeader(ctx, metadata.Pairs("HashSHA256", hashStr))
	}

	// Perform backup if in sync mode
	if s.backup.IsSyncMode() {
		var backupVal string
		switch req.Metric.Type {
		case format.Gauge:
			backupVal = strconv.FormatFloat(req.Metric.Value, 'f', -1, 64)
		case format.Counter:
			backupVal = strconv.FormatInt(req.Metric.Delta, 10)
		default:
			log.Error("Undefined metric type", zap.String("type", req.Metric.Type))
			return nil, status.Errorf(codes.Internal, `Undefined metric type: %s`, req.Metric.Type)
		}
		_ = s.backup.SaveToStruct(req.Metric.Type, req.Metric.Id, backupVal)

		s.backup.SaveToFile()
	}

	return &response, nil
}

// GetValue gets the value of a metric
func (s *MetricsServer) GetValue(ctx context.Context, req *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	const op = "grpc.metric.GetValue"

	log := s.logger.With(
		zap.String("op", op),
		zap.String("request_id", middleware.GetReqID(ctx)),
	)

	// Validate the metric request
	if req.Metric.Id == "" || req.Metric.Type == "" {
		log.Error(
			"Name or Type is empty!",
			zap.String("name", req.Metric.Id),
			zap.String("type", req.Metric.Type))
		return nil, status.Errorf(codes.NotFound, `Metric %s not found`, req.Metric.Id)
	}

	// Create a context with a timeout for the database operation
	databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Retrieve the metric value from storage
	value, errStor := s.storage.GetMetric(databaseCtx, req.Metric.Type, req.Metric.Id)
	if errors.Is(errStor, storage.ErrMetricNotFound) {
		log.Info(
			"Metric is not found",
			zap.String("name", req.Metric.Id),
			zap.String("type", req.Metric.Type))
		return nil, status.Errorf(codes.NotFound, `Metric %s not found`, req.Metric.Id)
	}
	if errStor != nil && !errors.Is(errStor, storage.ErrMetricNotFound) {
		log.Error("Failed to get metric", zap.Error(errStor))
		return nil, status.Errorf(codes.Internal, `Failed to get metric: %s`, req.Metric.Id)
	}

	var metricResponse pb.GetValueResponse

	// Parse the metric value based on its type
	switch req.Metric.Type {
	case format.Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Error("Failed to parse gauge value", zap.Error(err))
			return nil, status.Errorf(codes.InvalidArgument, `Failed to parse gauge value: %s`, value)
		}
		metricResponse.Metric = &pb.Metric{
			Id:    req.Metric.Id,
			Type:  req.Metric.Type,
			Value: val,
		}
	case format.Counter:
		val, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			log.Error("Failed to parse counter value", zap.Error(err))
			return nil, status.Errorf(codes.InvalidArgument, `Failed to parse counter value: %s`, value)
		}
		metricResponse.Metric = &pb.Metric{
			Id:    req.Metric.Id,
			Type:  req.Metric.Type,
			Delta: val,
		}
	default:
		log.Error("Undefined metric type", zap.String("type", req.Metric.Type))
		return nil, status.Errorf(codes.InvalidArgument, `Undefined metric type: %s`, req.Metric.Type)
	}

	// If a SHA256 key is provided, generate and set the hash header
	if s.conf.Key != "" {
		marshal, err := json.Marshal(metricResponse.Metric)
		if err != nil {
			log.Error("Error encoding response", zap.Error(err))
			return nil, status.Error(codes.Internal, `Error encoding response`)
		}
		hashStr := signature.GetHash(s.conf.Key, string(marshal), log)
		_ = grpc.SetHeader(ctx, metadata.Pairs("HashSHA256", hashStr))
	}

	return &metricResponse, nil
}

// Home is the home page handler
func (s *MetricsServer) Home(ctx context.Context, req *pb.HomeRequest) (*pb.HomeResponse, error) {
	const op = "grpc.metric.Home"

	log := s.logger.With(
		zap.String("op", op),
		zap.String("request_id", middleware.GetReqID(ctx)),
	)

	databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gauge, counter, err := s.storage.GetAllMetrics(databaseCtx)
	if err != nil {
		log.Error("Failed to get all metrics", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to get all metrics")
	}
	log.Info("Metrics received", zap.Any("gauge", gauge), zap.Any("counter", counter))

	var metrics []*pb.Metric

	for _, metric := range gauge {
		val, err := strconv.ParseFloat(metric[1], 64)
		if err != nil {
			log.Error("Failed to parse gauge value", zap.Error(err))
			return nil, status.Errorf(codes.InvalidArgument, "Failed to parse gauge value: %s", metric[1])
		}
		metrics = append(metrics, &pb.Metric{
			Id:    metric[0],
			Type:  format.Gauge,
			Value: val,
		})
	}

	for _, metric := range counter {
		val, err := strconv.ParseInt(metric[1], 0, 64)
		if err != nil {
			log.Error("Failed to parse counter value", zap.Error(err))
			return nil, status.Errorf(codes.InvalidArgument, "Failed to parse counter value: %s", metric[1])
		}
		metrics = append(metrics, &pb.Metric{
			Id:    metric[0],
			Type:  format.Counter,
			Delta: val,
		})
	}

	response := &pb.HomeResponse{
		Metrics: metrics,
	}

	if s.conf.Key != "" {
		marshal, err := json.Marshal(response)
		if err != nil {
			log.Error("Error encoding response", zap.Error(err))
			return nil, status.Error(codes.Internal, "Error encoding response")
		}
		hashStr := signature.GetHash(s.conf.Key, string(marshal), log)
		_ = grpc.SetHeader(ctx, metadata.Pairs("HashSHA256", hashStr))
	}

	return response, nil
}

// / UpdateBatch updates multiple metrics at once
func (s *MetricsServer) UpdateBatch(ctx context.Context, req *pb.UpdateBatchRequest) (*pb.UpdateBatchResponse, error) {
	const op = "handlers.updates.NewJSON"

	log := s.logger.With(
		zap.String("op", op),
		zap.String("request_id", middleware.GetReqID(ctx)),
	)

	var gauges [][]string
	var counters [][]string
	for _, metric := range req.Metrics {
		switch metric.Type {
		case format.Gauge:
			gauges = append(gauges, []string{metric.Id, strconv.FormatFloat(metric.Value, 'f', -1, 64)})
		case format.Counter:
			counters = append(counters, []string{metric.Id, strconv.FormatInt(metric.Delta, 10)})
		default:
			log.Error("Unknown metric type", zap.String("type", metric.Type))
			return nil, status.Errorf(codes.InvalidArgument, "Unknown metric type: %s", metric.Type)
		}

		err := backupHandler(log, s.backup, metric)
		if err != nil {
			log.Error("Cannot backup metric", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "Cannot backup metric")
		}
	}

	databaseCtx, cancel := context.WithTimeout(ctx, 11*time.Second)
	defer cancel()
	err := s.storage.UpdateBatch(databaseCtx, gauges, counters)

	if err != nil {
		log.Error("Failed to batch update", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to batch update")
	}

	response := &pb.UpdateBatchResponse{
		Metrics: req.Metrics,
	}

	if s.conf.Key != "" {
		marshal, err := json.Marshal(response)
		if err != nil {
			log.Error("Error encoding response", zap.Error(err))
			return nil, status.Error(codes.Internal, "Error encoding response")
		}
		hashStr := signature.GetHash(s.conf.Key, string(marshal), log)
		_ = grpc.SetHeader(ctx, metadata.Pairs("HashSHA256", hashStr))
	}

	return response, nil
}

// backupHandler handles the backup of a single metric.
// It takes a logger, backup handler, and the metric to be backed up.
// Returns an error if the metric type is undefined or if the backup fails.
func backupHandler(log *zap.Logger, backup *backuper.Buckuper, metric *pb.Metric) error {
	const op = "handlers.updates.backup"
	if backup.IsSyncMode() {
		var backupVal string
		switch metric.Type {
		case format.Gauge:
			backupVal = strconv.FormatFloat(metric.Value, 'f', -1, 64)
		case format.Counter:
			backupVal = strconv.FormatInt(metric.Delta, 10)
		default:
			log.Error("Undefined metric type", zap.String("type", metric.Type))
			return fmt.Errorf("%s: %s", op, "Undefined metric type")
		}
		_ = backup.SaveToStruct(metric.Type, metric.Id, backupVal)
		backup.SaveToFile()
	}
	return nil
}
