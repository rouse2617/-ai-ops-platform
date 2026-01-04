package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Client Prometheus 客户端别名
type Client = PrometheusClient

// PrometheusClient Prometheus 客户端
type PrometheusClient struct {
	client v1.API
	addr   string
}

// NewPrometheusClient 创建 Prometheus 客户端
func NewPrometheusClient(addr string) (*PrometheusClient, error) {
	client, err := api.NewClient(api.Config{
		Address: addr,
	})
	if err != nil {
		return nil, err
	}

	return &PrometheusClient{
		client: v1.NewAPI(client),
		addr:   addr,
	}, nil
}

// QueryRange 范围查询
func (pc *PrometheusClient) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (model.Matrix, error) {
	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	result, warnings, err := pc.client.QueryRange(ctx, query, r)
	if err != nil {
		return nil, err
	}

	if len(warnings) > 0 {
		fmt.Printf("Prometheus warnings: %v\n", warnings)
	}

	matrix, ok := result.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return matrix, nil
}

// QueryInstant 即时查询
func (pc *PrometheusClient) QueryInstant(ctx context.Context, query string, ts time.Time) (model.Vector, error) {
	result, warnings, err := pc.client.Query(ctx, query, ts)
	if err != nil {
		return nil, err
	}

	if len(warnings) > 0 {
		fmt.Printf("Prometheus warnings: %v\n", warnings)
	}

	vector, ok := result.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return vector, nil
}

// GetMetricMetadata 获取指标元数据
func (pc *PrometheusClient) GetMetricMetadata(ctx context.Context) (map[string][]v1.Metadata, error) {
	return pc.client.Metadata(ctx, "", "")
}

// GetLabelValues 获取标签值
func (pc *PrometheusClient) GetLabelValues(ctx context.Context, label string, start, end time.Time) (model.LabelValues, v1.Warnings, error) {
	return pc.client.LabelValues(ctx, label, nil, start, end)
}
