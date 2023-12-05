package utils

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

type AuthorizerContext struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	ExecutionContext
}

type ExecutionContext struct {
	AWSRegion string
	FirstName string
	LastName  string
	Email     string
	TenantID  string
	UserRole  string
	UserID    string
}

func BuildExecutionEnvironment(request events.APIGatewayProxyRequest) (*AuthorizerContext, *ExecutionContext, error) {
	var auth AuthorizerContext
	var exec ExecutionContext

	authorizer := request.RequestContext.Authorizer

	tenantID, ok := authorizer["tenantId"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("tenantId not found in context")
	}
	exec.TenantID = tenantID // Removed extra 'c' at the end of tenantID

	awsRegion, ok := authorizer["awsRegion"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("awsRegion not found in context")
	}
	exec.AWSRegion = awsRegion

	accessKeyId, ok := authorizer["accessKeyId"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("accessKeyId not found in context")
	}
	auth.AccessKeyID = accessKeyId

	secretAccessKey, ok := authorizer["secretAccessKey"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("secretAccessKey not found in context")
	}
	auth.SecretAccessKey = secretAccessKey

	sessionToken, ok := authorizer["sessionToken"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("sessionToken not found in context")
	}
	auth.SessionToken = sessionToken

	firstName, ok := authorizer["firstName"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("firstName not found in context")
	}
	exec.FirstName = firstName

	lastName, ok := authorizer["lastName"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("lastName not found in context")
	}
	exec.LastName = lastName

	email, ok := authorizer["email"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("email not found in context")
	}
	exec.Email = email

	userRole, ok := authorizer["userRole"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("userRole not found in context")
	}
	exec.UserRole = userRole

	userID, ok := authorizer["userId"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("userId not found in context")
	}
	exec.UserID = userID

	contextLogger := NewContextLogger(&exec)
	contextLogger.InfoLog("Authorizer details fetched successfully", Fields{"execution_context": exec})

	return &auth, &exec, nil
}
