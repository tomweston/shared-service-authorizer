package utils

import (
	"context"
	"os"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	"github.com/aws/aws-lambda-go/events"
)

const baseMetricName = "vantagea."

type APIMetrics interface {
	RecordSuccess(ctx context.Context, request events.APIGatewayProxyRequest, metricName string)
	RecordError(ctx context.Context, request events.APIGatewayProxyRequest, metricName string)
}

type DataDogMetrics struct{}

func (dd *DataDogMetrics) RecordSuccess(ctx context.Context, request events.APIGatewayProxyRequest, metricName string) {
	tags := getCommonTags(ctx, request)
	ddlambda.Metric(baseMetricName+metricName+".success", 1.0, tags...)
}

func (dd *DataDogMetrics) RecordError(ctx context.Context, request events.APIGatewayProxyRequest, metricName string) {
	tags := getCommonTags(ctx, request)
	ddlambda.Metric(baseMetricName+metricName+".errors", 1.0, tags...)
}

// getCommonTags extracts common tags from the context and request.
func getCommonTags(ctx context.Context, request events.APIGatewayProxyRequest) []string {
	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	functionVersion := os.Getenv("AWS_LAMBDA_FUNCTION_VERSION")
	functionExecutionEnv := os.Getenv("AWS_EXECUTION_ENV")
	functionMemorySize := os.Getenv("AWS_LAMBDA_FUNCTION_MEMORY_SIZE")
	region := os.Getenv("AWS_REGION")
	return []string{
		"tenant:" + request.RequestContext.Authorizer["tenantId"].(string),
		"path:" + request.Path,
		"method:" + request.HTTPMethod,
		"stage:" + request.RequestContext.Stage,
		"request_id:" + request.RequestContext.RequestID,
		"source_ip:" + request.RequestContext.Identity.SourceIP,
		"function_name:" + functionName,
		"function_version:" + functionVersion,
		"function_execution_env:" + functionExecutionEnv,
		"function_memory_size:" + functionMemorySize,
		"aws_region:" + region,
	}
}

// NewDataDogMetrics returns an instance of DataDogMetrics.
func NewDataDogMetrics() APIMetrics {
	return &DataDogMetrics{}
}
