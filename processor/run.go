package processor

import (
	"context"
	"fmt"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
)

type promClient interface {
	Query(ctx context.Context, query string, ts time.Time, opts ...promv1.Option) (prommodel.Value, promv1.Warnings, error)
}

type RunProcessorOptions struct {
	Timeout     int
	Concurrency int
	ThanosURL   string
	Quantile    string
	Mode        string
	LimitMargin string
}

type RunProcessor struct {
	Options RunProcessorOptions
	Client  promClient
}

func (p *RunProcessor) Type() string {
	return "RunProcessor"
}

type prometheusMetrics struct {
	RequestCPU map[string]float64
	LimitCPU   map[string]float64
	RequestMem map[string]float64
	LimitMem   map[string]float64
}

type calculatedMetrics struct {
	CPUUtilizationRatio    float64
	MemoryUtilizationRatio float64
	CPUOverProvision       float64
	MemoryOverProvision    float64
}

const (
	podCPURequest    = `quantile_over_time(%s, node_namespace_pod_container:container_cpu_usage_seconds_total:%s{pod="%s", container!=""}[1w])`
	podCPULimit      = `max_over_time(node_namespace_pod_container:container_cpu_usage_seconds_total:%s{pod="%s", container!=""}[1w]) * %s`
	podMemoryRequest = `quantile_over_time(%s, container_memory_working_set_bytes{pod="%s", container!=""}[1w]) / 1024 / 1024`
	podMemoryLimit   = `(max_over_time(container_memory_working_set_bytes{pod="%s", container!=""}[1w]) / 1024 / 1024) * %s`
)

func newPromClient(prometheusURL string) (promClient, error) {
	client, err := promapi.NewClient(promapi.Config{Address: prometheusURL})
	if err != nil {
		return nil, fmt.Errorf("error creating Prometheus client: %w", err)
	}
	return promv1.NewAPI(client), nil
}

func queryPrometheus(ctx context.Context, client promClient, query string, ts time.Time) (prommodel.Vector, error) {
	result, warnings, err := client.Query(ctx, query, ts)
	if err != nil {
		return nil, fmt.Errorf("error querying Prometheus: %w", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("Prometheus query warnings: %v\n", warnings)
	}

	vector, ok := result.(prommodel.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from Prometheus: %T", result)
	}

	return vector, nil
}

func queryStatistic(ctx context.Context, client promClient, query string, ts time.Time) (map[string]float64, error) {
	output := make(map[string]float64)

	vector, err := queryPrometheus(ctx, client, query, ts)
	if err != nil {
		return output, err
	}

	// Iterate over the vector and extract metrics
	for _, sample := range vector {
		containerName := string(sample.Metric["container"])
		output[containerName] = float64(sample.Value)
	}

	return output, nil
}

func (o *RunProcessorOptions) queryPrometheus(ctx context.Context, client promClient, podName string) (prometheusMetrics, error) {
	now := time.Now()
	var err error
	output := prometheusMetrics{}

	output.RequestCPU, err = queryStatistic(ctx, client, fmt.Sprintf(podCPURequest, o.Quantile, o.Mode, podName), now)
	if err != nil {
		return output, err
	}
	output.LimitCPU, err = queryStatistic(ctx, client, fmt.Sprintf(podCPULimit, o.Mode, podName, o.LimitMargin), now)
	if err != nil {
		return output, err
	}
	output.RequestMem, err = queryStatistic(ctx, client, fmt.Sprintf(podMemoryRequest, o.Quantile, podName), now)
	if err != nil {
		return output, err
	}
	output.LimitMem, err = queryStatistic(ctx, client, fmt.Sprintf(podMemoryLimit, podName, o.LimitMargin), now)
	if err != nil {
		return output, err
	}

	return output, nil
}

func calculateMetrics(metrics prometheusMetrics) calculatedMetrics {
	calc := calculatedMetrics{}

	// Sum the CPU and memory values for all containers
	var totalCPURequest, totalCPULimit, totalMemoryRequest, totalMemoryLimit float64

	for _, cpuRequest := range metrics.RequestCPU {
		totalCPURequest += cpuRequest
	}
	for _, cpuLimit := range metrics.LimitCPU {
		totalCPULimit += cpuLimit
	}
	for _, memRequest := range metrics.RequestMem {
		totalMemoryRequest += memRequest
	}
	for _, memLimit := range metrics.LimitMem {
		totalMemoryLimit += memLimit
	}

	// Calculate CPU and Memory utilization ratios
	if totalCPULimit > 0 {
		calc.CPUUtilizationRatio = totalCPURequest / totalCPULimit
	}
	if totalMemoryLimit > 0 {
		calc.MemoryUtilizationRatio = totalMemoryRequest / totalMemoryLimit
	}

	// Over-provisioning ratios: how much limit exceeds request
	if totalCPURequest > 0 {
		calc.CPUOverProvision = totalCPULimit / totalCPURequest
	}
	if totalMemoryRequest > 0 {
		calc.MemoryOverProvision = totalMemoryLimit / totalMemoryRequest
	}

	return calc
}

func (p *RunProcessor) Run(podName string) {
	ctx := context.Background()

	metrics, err := p.Options.queryPrometheus(ctx, p.Client, podName)
	if err != nil {
		fmt.Printf("Error querying pod metrics: %v\n", err)
		return
	}

	calculated := calculateMetrics(metrics)

	fmt.Printf("Pod: %s\n", podName)
	fmt.Printf("CPU Request: %v\n", metrics.RequestCPU)
	fmt.Printf("CPU Limit: %v\n", metrics.LimitCPU)
	fmt.Printf("Memory Request: %v\n", metrics.RequestMem)
	fmt.Printf("Memory Limit: %v\n", metrics.LimitMem)
	fmt.Printf("CPU Utilization Ratio: %.2f\n", calculated.CPUUtilizationRatio)
	fmt.Printf("Memory Utilization Ratio: %.2f\n", calculated.MemoryUtilizationRatio)
	fmt.Printf("CPU Over-provision Ratio: %.2f\n", calculated.CPUOverProvision)
	fmt.Printf("Memory Over-provision Ratio: %.2f\n", calculated.MemoryOverProvision)

	fmt.Println("Timeout:", p.Options.Timeout)
	fmt.Println("Concurrency:", p.Options.Concurrency)
}

func NewRunProcessor(options RunProcessorOptions) (*RunProcessor, error) {
	client, err := newPromClient(options.ThanosURL)
	if err != nil {
		return nil, err
	}
	return &RunProcessor{
		Options: options,
		Client:  client,
	}, nil
}
