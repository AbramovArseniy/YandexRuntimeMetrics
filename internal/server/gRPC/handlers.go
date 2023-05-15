package grpc_server

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/loggers"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/myerrors"
	pb "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/proto"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/config"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/database"
	filestorage "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/fileStorage"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/storage"
	"github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/types"
)

// Server has gRPC server info
type MetricServer struct {
	pb.UnimplementedMetricsServer

	Storage       storage.Storage
	StorageType   types.StorageType
	Key           string
	Addr          string
	Debug         bool
	CryptoKey     *rsa.PrivateKey
	TrustedSubnet string
}

// NewServer creates new Server
func NewMetricServer(cfg config.Config) *MetricServer {
	var (
		storage     storage.Storage
		storageType types.StorageType
	)

	var cryptoKey *rsa.PrivateKey
	if cfg.CryptoKeyFile != "" {
		file, err := os.OpenFile(cfg.CryptoKeyFile, os.O_RDONLY, 0777)
		if err != nil {
			loggers.ErrorLogger.Println("error while opening crypto key file:", err)
			cryptoKey = nil
		} else {
			cryptoKeyByte, err := io.ReadAll(file)
			if err != nil {
				loggers.ErrorLogger.Println("error while reading crypto key file:", err)
				cryptoKey = nil
			}
			cryptoKey, err = x509.ParsePKCS1PrivateKey(cryptoKeyByte)
			if err != nil {
				loggers.ErrorLogger.Println("error while parsing crypto key:", err)
				cryptoKey = nil
			}
		}
	}
	if cfg.Database == nil {
		fs := filestorage.NewFileStorage(cfg)
		fs.SetFileStorage()
		storage = fs
		storageType = types.StorageTypeFile
	} else {
		storage = database.NewDatabase(cfg.Database)
		storageType = types.StorageTypeDB
	}
	return &MetricServer{
		Addr:          cfg.Address,
		Debug:         cfg.Debug,
		Key:           cfg.HashKey,
		Storage:       storage,
		StorageType:   storageType,
		CryptoKey:     cryptoKey,
		TrustedSubnet: cfg.TrustedSubnet,
	}
}

func (s *MetricServer) CheckRequestSubnetInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var strIP string
	if s.TrustedSubnet == "" {
		return handler(ctx, req)
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("X-Real-IP")
		if len(values) > 0 {
			strIP = values[0]
		}
	}
	if len(strIP) == 0 {
		return nil, status.Error(codes.PermissionDenied, "the client's IP is not in the trusted subnet")
	}
	_, IPNet, err := net.ParseCIDR(s.TrustedSubnet)
	if err != nil {
		loggers.ErrorLogger.Println("error while parsing trusted subnet CIDR:", err)
		return handler(ctx, req)
	}
	clientIP := net.ParseIP(strIP)
	if clientIP == nil {
		loggers.ErrorLogger.Println("error while getting client's IP:", err)
		return handler(ctx, req)
	}
	if !IPNet.Contains(clientIP) {
		return nil, status.Error(codes.PermissionDenied, "the client's IP is not in the trusted subnet")
	}
	return handler(ctx, req)
}

func (s *MetricServer) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	var response pb.UpdateMetricResponse
	var m = types.Metrics{
		ID:    in.Metric.Id,
		MType: in.Metric.Mtype,
	}
	switch in.Metric.Mtype {
	case "counter":
		var delta int64 = in.Metric.Delta
		m.Delta = &delta
		curval, err := s.Storage.GetMetric(m, s.Key)
		if err == nil {
			in.Metric.Delta = *curval.Delta + in.Metric.Delta
		}
		response.Metric = &pb.Metric{
			Id:    m.ID,
			Mtype: m.MType,
			Delta: *m.Delta + *curval.Delta,
		}
	case "gauge":
		var value float64 = in.Metric.Value
		m.Value = &value
		response.Metric = &pb.Metric{
			Id:    m.ID,
			Mtype: m.MType,
			Value: *m.Value,
		}
	default:
		return nil, status.Error(codes.Unimplemented, "wrong metric type")
	}
	err := s.Storage.SaveMetric(m, s.Key)
	if err != nil {
		loggers.ErrorLogger.Println("error while saving metric:", err)
		return nil, status.Error(codes.Internal, "error while saving metric")
	}
	return &response, nil
}

func (s *MetricServer) UpdateManyMetrics(ctx context.Context, in *pb.UpdateManyMetricsRequest) (*pb.UpdateManyMetricsResponse, error) {
	var response pb.UpdateManyMetricsResponse
	var m []types.Metrics = make([]types.Metrics, len(in.Metrics))
	for i, metric := range in.Metrics {
		switch metric.Mtype {
		case "counter":
			var delta int64 = metric.Delta
			m[i] = types.Metrics{
				ID:    metric.Id,
				MType: metric.Mtype,
				Delta: &delta,
			}
			curval, err := s.Storage.GetMetric(m[i], s.Key)
			if err == nil {
				in.Metrics[i].Delta = *curval.Delta + metric.Delta
			}
		case "gauge":
			var value float64 = metric.Value
			m[i] = types.Metrics{
				ID:    metric.Id,
				MType: metric.Mtype,
				Value: &value,
			}
		default:
			return nil, status.Error(codes.Unimplemented, "wrong metric type")
		}
	}
	s.Storage.SaveManyMetrics(m, s.Key)
	response = pb.UpdateManyMetricsResponse{
		Metrics: in.Metrics,
	}
	return &response, nil
}

func (s *MetricServer) GetMetric(ctx context.Context, in *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	var response pb.GetMetricResponse
	var m = types.Metrics{
		ID:    in.Metric.Id,
		MType: in.Metric.Mtype,
	}
	curval, err := s.Storage.GetMetric(m, s.Key)
	if errors.Is(err, myerrors.ErrTypeNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, myerrors.ErrTypeNotImplemented) {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}
	switch in.Metric.Mtype {
	case "counter":
		response.Metric = &pb.Metric{
			Id:    m.ID,
			Mtype: m.MType,
			Delta: *m.Delta + *curval.Delta,
		}
	case "gauge":
		response.Metric = &pb.Metric{
			Id:    m.ID,
			Mtype: m.MType,
			Value: *m.Value,
		}
	}
	return &response, nil
}

func (s *MetricServer) GetAllMetrics(ctx context.Context, in *pb.GetAllMetricsRequest) (*pb.GetAllMetricsResponse, error) {
	var response pb.GetAllMetricsResponse
	metrics, err := s.Storage.GetAllMetrics()
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot get values of metrics")
	}
	var responseMetrics = make([]*pb.Metric, len(metrics))
	for i, m := range metrics {
		switch m.MType {
		case "counter":
			responseMetrics[i] = &pb.Metric{
				Id:    m.ID,
				Mtype: m.MType,
				Delta: *m.Delta,
			}
		case "gauge":
			responseMetrics[i] = &pb.Metric{
				Id:    m.ID,
				Mtype: m.MType,
				Value: *m.Value,
			}
		}
	}
	response.Metrics = responseMetrics
	return &response, nil
}

func (s *MetricServer) PingDatabase(ctx context.Context, _ *pb.PingDatabaseRequest) (*pb.PingDatabaseResponse, error) {
	if err := s.Storage.Check(); err != nil || s.StorageType != types.StorageTypeDB {
		return nil, status.Error(codes.Internal, "failed to ping database")
	}
	return nil, nil
}
