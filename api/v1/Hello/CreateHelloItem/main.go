package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/segmentio/ksuid"
	"github.com/tomweston/shared-service-authorizer/utils"
)

var metrics utils.APIMetrics

func init() {
	metrics = utils.NewDataDogMetrics()
}

// GetIPInfo gets the IP info for the requester
func GetIPInfo(request events.APIGatewayProxyRequest, auth *utils.AuthorizerContext, execContext *utils.ExecutionContext) *ipinfo.Core {
	contextLogger := utils.NewContextLogger(execContext)
	fields := utils.NewFields()

	token := os.Getenv("IPINFO_TOKEN")
	client := ipinfo.NewClient(nil, nil, token)
	ipInfo, err := client.GetIPInfo(net.ParseIP(request.RequestContext.Identity.SourceIP))

	if err != nil {
		fields["error"] = err
		contextLogger.ErrorLog("Failed to get IP Info", fields)
		metrics.RecordError(context.Background(), request, "ipinfo")
	}

	metrics.RecordSuccess(context.Background(), request, "ipinfo")
	return ipInfo
}

// CreateHelloItem creates a new item in the SharedServices DynamoDB table
func CreateHelloItem(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	auth, execContext, err := utils.BuildExecutionEnvironment(request)
	if err != nil {
		return formatErrorResponse(500, "Internal Server Error"), err
	}

	// Initialize your DynamoDB service client and IPInfo client
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(auth.AWSRegion),
	}))
	svc := dynamodb.New(sess)

	ipinfoData := GetIPInfo(request, auth, execContext)

	contextLogger := utils.NewContextLogger(execContext)
	metrics := utils.NewDataDogMetrics()

	id := ksuid.New().String()
	item := map[string]*dynamodb.AttributeValue{
		"PK": {
			S: aws.String("TENANT#" + execContext.TenantID),
		},
		"SK": {
			S: aws.String("HELLO#" + id),
		},
		"ID": {
			S: aws.String(id),
		},
		"Requester": {
			S: aws.String(execContext.FirstName + " " + execContext.LastName),
		},
		"City": {
			S: aws.String(ipinfoData.City),
		},
		"Postal": {
			S: aws.String(ipinfoData.Postal),
		},
		"Region": {
			S: aws.String(ipinfoData.Region),
		},
		"Timezone": {
			S: aws.String(ipinfoData.Timezone),
		},
		"Country": {
			S: aws.String(ipinfoData.Country),
		},
		"CountryName": {
			S: aws.String(ipinfoData.CountryName),
		},
		"CountryFlag": {
			S: aws.String(ipinfoData.CountryFlag.Emoji),
		},
		"CountryFlagURL": {
			S: aws.String(ipinfoData.CountryFlagURL),
		},
		"CountryCurrencyCode": {
			S: aws.String(ipinfoData.CountryCurrency.Code),
		},
		"CountryCurrencySymbol": {
			S: aws.String(ipinfoData.CountryCurrency.Symbol),
		},
		"ContinentCode": {
			S: aws.String(ipinfoData.Continent.Code),
		},
		"ContinentName": {
			S: aws.String(ipinfoData.Continent.Name),
		},
		"IsEU": {
			BOOL: aws.Bool(ipinfoData.IsEU),
		},
		"Location": {
			S: aws.String(ipinfoData.Location),
		},
		"Org": {
			S: aws.String(ipinfoData.Org),
		},
	}

	putInput := &dynamodb.PutItemInput{
		TableName: aws.String("SharedServices"),
		Item:      item,
	}

	_, err = svc.PutItem(putInput)
	if err != nil {
		fields := utils.NewFields()
		fields["error"] = err
		contextLogger.ErrorLog("Failed to put item to DynamoDB", fields)
		metrics.RecordError(ctx, request, "createHelloItem")
		return formatErrorResponse(500, "Internal Server Error"), err
	}
	metrics.RecordSuccess(ctx, request, "createHelloItem")

	// Convert the single item map to a slice of maps
	itemSlice := []map[string]*dynamodb.AttributeValue{item}

	// Format the item slice to JSON using your function
	itemJSON, err := formatItemsToJSON(itemSlice)
	if err != nil {
		return formatErrorResponse(500, "Internal Server Error"), err
	}

	// Return the JSON-encoded item as the response body
	return events.APIGatewayProxyResponse{
		Body:       itemJSON,
		StatusCode: 200,
	}, nil
}

func formatErrorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	errorResponse := map[string]string{"error": message}
	errorJSON, err := json.Marshal(errorResponse)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "An unexpected error occurred: %v"}`, err),
			StatusCode: statusCode,
		}
	}
	return events.APIGatewayProxyResponse{
		Body:       string(errorJSON),
		StatusCode: statusCode,
	}
}

func formatItemsToJSON(items []map[string]*dynamodb.AttributeValue) (string, error) {
	var itemInterfaces []map[string]interface{}
	for _, item := range items {
		itemInterface := map[string]interface{}{}
		err := dynamodbattribute.UnmarshalMap(item, &itemInterface)
		if err != nil {
			return "", err
		}
		itemInterfaces = append(itemInterfaces, itemInterface)
	}
	jsonString, err := json.Marshal(itemInterfaces)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

func main() {
	lambda.Start(CreateHelloItem)
}
