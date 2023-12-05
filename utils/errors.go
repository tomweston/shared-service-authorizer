package utils

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func FormatErrorResponse(statusCode int, messages ...string) events.APIGatewayProxyResponse {
	errorResponse := map[string][]string{"errors": messages}
	errorJSON, err := json.Marshal(errorResponse)
	if err != nil {
		rawError := map[string][]string{"errors": {fmt.Sprintf("An unexpected error occurred: %v", err)}}
		rawErrorJSON, _ := json.Marshal(rawError)
		return events.APIGatewayProxyResponse{
			Body:       string(rawErrorJSON),
			StatusCode: statusCode,
		}
	}
	return events.APIGatewayProxyResponse{
		Body:       string(errorJSON),
		StatusCode: statusCode,
	}
}
