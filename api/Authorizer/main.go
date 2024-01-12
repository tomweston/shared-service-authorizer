////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////
//// 	This Shared Service Authorizer is inspired by the AWS SaaS Factory Serverless SaaS Identity and Isolation Patterns
////
//// 	Reference PDF:
//// 	./docs/SaaS_tenant_isolation_patterns.pdf
////
//// 	Reference Implementation Repo:
////	https://github.com/aws-samples/aws-saas-factory-ref-solution-serverless-saas/
////	Reference Code:
////	https://github.com/aws-samples/aws-saas-factory-ref-solution-serverless-saas/blob/main/server/Resources/shared_service_authorizer.py
////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////
////	The Shared Service Authorizer is a Lambda function that is invoked by API Gateway to authorize requests to the Shared Services API.
////
////	It is responsible for:
////		- Validating the JWT token
////		- Generating STS credentials for the tenant
////		- Generating an IAM policy for the tenant
////		- Returning the policy and STS credentials to API Gateway
////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"
)

// UserRoles enumeration
var UserRoles = struct {
	SYSTEM_ADMIN     string
	CUSTOMER_SUPPORT string
	TENANT_ADMIN     string
	TENANT_USER      string
}{
	"SystemAdmin",
	"CustomerSupport",
	"TenantAdmin",
	"TenantUser",
}

func isTenantAdmin(userRole string) bool {
	return userRole == UserRoles.TENANT_ADMIN
}

func isSystemAdmin(userRole string) bool {
	return userRole == UserRoles.SYSTEM_ADMIN
}

// TODO: This function will be used when we have a SaaS provider role (e.g. SystemAdmin or CustomerSupport)
func isSaaSProvider(userRole string) bool {
	return userRole == UserRoles.SYSTEM_ADMIN || userRole == UserRoles.CUSTOMER_SUPPORT
}

func isTenantUser(userRole string) bool {
	return userRole == UserRoles.TENANT_USER
}

func GetPolicyForUser(userRole, serviceIdentifier, tenantID, region, awsAccountID string) string {
	var iamPolicy string

	if isSystemAdmin(userRole) {
		iamPolicy = GetPolicyForSystemAdmin(region, awsAccountID)
	} else if isTenantAdmin(userRole) {
		iamPolicy = GetPolicyForTenantAdmin(tenantID, serviceIdentifier, region, awsAccountID)
	} else if isTenantUser(userRole) {
		iamPolicy = GetPolicyForTenantUser(tenantID, region, awsAccountID)
	}

	// log.Printf("IAM Policy: %s", iamPolicy)
	return iamPolicy
}

type APIGatewayErrorResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"requestID"`
}

func GetPolicyForSystemAdmin(region, awsAccountID string) string {
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Action": []string{
					"dynamodb:UpdateItem",
					"dynamodb:GetItem",
					"dynamodb:PutItem",
					"dynamodb:Query",
					"dynamodb:Scan",
					"dynamodb:DeleteItem",
					"dynamodb:BatchWriteItem",
					"dynamodb:BatchGetItem",
				},
				"Resource": []string{
					fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/*", region, awsAccountID),
				},
			},
		},
	}

	iamPolicy, _ := json.Marshal(policy)
	return string(iamPolicy)
}

func GetPolicyForTenantAdmin(tenantID, serviceIdentifier, region, awsAccountID string) string {
	var policy map[string]interface{}

	if serviceIdentifier == "SharedServices" {
		policy = map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect": "Allow",
					"Action": []string{
						"dynamodb:UpdateItem",
						"dynamodb:GetItem",
						"dynamodb:PutItem",
						"dynamodb:Query",
						"dynamodb:Scan",
						"dynamodb:DeleteItem",
						"dynamodb:BatchWriteItem",
						"dynamodb:BatchGetItem",
					},
					"Resource": []string{
						fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/SharedServices", region, awsAccountID),
					},
					"Condition": map[string]interface{}{
						"ForAllValues:StringEquals": map[string]interface{}{
							"dynamodb:LeadingKeys": []string{
								"TENANT#" + tenantID,
							},
						},
					},
				},
				{
					"Effect": "Allow",
					"Action": []string{
						"dynamodb:UpdateItem",
						"dynamodb:GetItem",
						"dynamodb:PutItem",
						"dynamodb:Query",
						"dynamodb:Scan",
						"dynamodb:DeleteItem",
						"dynamodb:BatchWriteItem",
						"dynamodb:BatchGetItem",
					},
					"Resource": []string{
						fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/SharedServices", region, awsAccountID),
					},
				},
			},
		}
	} else {
		// This is where we would handle DedicatedTenantServices
	}

	iamPolicy, _ := json.Marshal(policy)
	return string(iamPolicy)
}

func GetPolicyForTenantUser(tenantID, region, awsAccountID string) string {
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Action": []string{
					"dynamodb:UpdateItem",
					"dynamodb:GetItem",
					"dynamodb:PutItem",
					"dynamodb:Query",
					"dynamodb:Scan",
					"dynamodb:DeleteItem",
					"dynamodb:BatchWriteItem",
					"dynamodb:BatchGetItem",
				},
				"Resource": []string{
					fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/SharedServices", region, awsAccountID),
				},
				"Condition": map[string]interface{}{
					"ForAllValues:StringLike": map[string]interface{}{
						"dynamodb:LeadingKeys": []string{
							"TENANT#" + tenantID,
						},
					},
				},
			},
			{
				"Effect": "Allow",
				"Action": []string{
					"dynamodb:UpdateItem",
					"dynamodb:GetItem",
					"dynamodb:PutItem",
					"dynamodb:Query",
					"dynamodb:Scan",
					"dynamodb:DeleteItem",
					"dynamodb:BatchWriteItem",
					"dynamodb:BatchGetItem",
				},
				"Resource": []string{
					fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/SharedServices", region, awsAccountID),
				},
				"Condition": map[string]interface{}{
					"ForAllValues:StringLike": map[string]interface{}{
						"dynamodb:LeadingKeys": []string{
							"TENANT#" + tenantID,
						},
					},
				},
			},
		},
	}

	iamPolicy, _ := json.Marshal(policy)
	return string(iamPolicy)
}

type CognitoJWTClaim struct {
	Region     string          `json:"custom:region"`
	Subject    string          `json:"sub"`
	LastName   string          `json:"custom:lastName"`
	Issuer     string          `json:"iss"`
	TenantID   string          `json:"custom:tenantId"`
	FirstName  string          `json:"custom:firstName"`
	UserRole   string          `json:"custom:userRole"`
	Email      string          `json:"email"`
	Expiration jwt.NumericDate `json:"exp"`
	IssuedAt   int64           `json:"iat"`
	Audience   string          `json:"aud"`
	jwt.Claims
}

func (c *CognitoJWTClaim) GetIssuedAt() (*jwt.NumericDate, error) {
	iat := jwt.NewNumericDate(time.Unix(c.IssuedAt, 0))
	return iat, nil
}

func (c *CognitoJWTClaim) GetExpirationTime() (*jwt.NumericDate, error) {
	if c.Expiration == struct{ time.Time }{} {
		return nil, errors.New("expiration is not set")
	}
	exp := jwt.NewNumericDate(time.Unix(c.Expiration.Unix(), 0))
	return exp, nil
}

func (c *CognitoJWTClaim) GetAudience() (jwt.ClaimStrings, error) {
	aud := jwt.ClaimStrings{c.Audience}
	return aud, nil
}

func (c *CognitoJWTClaim) GetNotBefore() (*jwt.NumericDate, error) {
	// Cognito does not set NotBefore
	return nil, nil
}

func (c *CognitoJWTClaim) GetIssuer() (string, error) {
	return c.Issuer, nil
}

func (c *CognitoJWTClaim) GetSubject() (string, error) {
	return c.Subject, nil
}

type Response events.APIGatewayCustomAuthorizerResponse

func assumeRole(sess *session.Session, roleArn, roleSessionName, policy string) (*sts.AssumeRoleOutput, error) {
	svc := sts.New(sess)
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(roleSessionName),
		Policy:          aws.String(policy),
	}

	result, err := svc.AssumeRole(input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type JSONError struct {
	Message string `json:"message"`
}

func (e JSONError) Error() string {
	return e.Message
}

func validateJWT(authToken string) (CognitoJWTClaim, error) {
	log.Printf("Validating JWT: %s", authToken)
	var claims CognitoJWTClaim
	parser := jwt.Parser{}
	_, _, err := parser.ParseUnverified(authToken, &claims)
	if err != nil {
		return claims, JSONError{Message: "Failed to parse unverified token"}
	}

	jwksURL := claims.Issuer + "/.well-known/jwks.json"
	jwks, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return claims, JSONError{Message: "Failed to fetch JWKS"}
	}

	token, err := jwt.ParseWithClaims(authToken, &claims, func(token *jwt.Token) (interface{}, error) {
		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, JSONError{Message: "Malformed token"}
		}

		key, found := jwks.LookupKeyID(keyID)
		if !found {
			return nil, JSONError{Message: "Invalid key ID"}
		}

		var pubkey interface{}
		err := key.Raw(&pubkey)
		if err != nil {
			return nil, JSONError{Message: "Invalid key type"}
		}

		return pubkey, nil
	})

	if !token.Valid {
		return claims, JSONError{Message: "Token is not valid"}
	}

	if time.Unix(claims.Expiration.Unix(), 0).Before(time.Now()) {
		return claims, JSONError{Message: "Token has expired"}
	}

	if err != nil {
		return claims, JSONError{Message: "Failed to parse token"}
	}

	return claims, nil
}

func getUsageIdentifierKey(tier string) (string, error) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"), // TODO: Get this from environment variable
	})
	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		return "", err
	}

	svc := apigateway.New(sess)

	params := &apigateway.GetApiKeysInput{
		IncludeValues: aws.Bool(true),
		NameQuery:     aws.String(tier + "Key"),
	}

	result, err := svc.GetApiKeys(params)
	if err != nil {
		log.Printf("Error getting API keys: %v", err)
		return "", err
	}

	if len(result.Items) > 0 {
		return *result.Items[0].Value, nil
	}

	return "", errors.New("usage identifier key not found")
}

func getAWSAccountID() (string, error) {
	sess, err := session.NewSession()
	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		return "", err
	}

	svc := sts.New(sess)
	result, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Printf("Error getting AWS account ID: %v", err)
		return "", err
	}

	return *result.Account, nil
}

func Handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (Response, error) {
	requestID := event.RequestContext.RequestID

	log.Printf("Starting Shared Service Authorizer")
	log.Printf("Request ID: %s Authorizer Request: %v", requestID, event)

	// log.Printf("Request ID: %s Method: %s Path: %s Headers: %s", requestID, event.MethodArn, event.Path, event.Headers["Authorization"])

	policy := Response{
		PrincipalID: "user",
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version:   "2012-10-17",
			Statement: []events.IAMPolicyStatement{},
		},
		Context: map[string]interface{}{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*",
			"Access-Control-Allow-Headers": "*",
			"Content-Type":                 "*/*",
		},
	}

	// Extract case-insensitive "Authorization" header
	var authorizationHeader string
	for k, v := range event.Headers {
		if strings.EqualFold(k, "Authorization") {
			authorizationHeader = v
			break
		}
	}

	log.Printf("Extracted Authorization Header Value: %v", authorizationHeader)

	claims, err := validateJWT(authorizationHeader)

	if err != nil {
		log.Printf("Request ID: %s Rejected: %v", requestID, err)
		response := APIGatewayErrorResponse{
			Message:   fmt.Sprintf("Rejected: %v", err),
			RequestID: requestID,
		}
		responseJSON, _ := json.Marshal(response)
		return policy, fmt.Errorf(string(responseJSON))
	}

	awsAccountID, err := getAWSAccountID()
	if err != nil {
		log.Printf("Request ID: %s Error getting AWS account ID: %v", requestID, err)
		return policy, err
	}

	// A map of region aliases to actual region names
	var region string
	switch claims.Region {
	case "eu1":
		region = "eu-west-2"
	case "us1":
		region = "us-east-1"
	case "ap1":
		region = "ap-southeast-1"
	default:
		log.Fatalf("Unexpected region: %s", claims.Region)
	}

	// TODO: Determine serviceIdentifier from ServiceIdentifier (e.g. SharedServices, DedicatedTenantServices) by looking up the value in the tenantDetails table

	iamPolicy := GetPolicyForUser(claims.UserRole, "SharedServices", claims.TenantID, region, awsAccountID)

	roleArn := "arn:aws:iam::" + awsAccountID + ":role/AuthorizerAccessRole"

	policy.PolicyDocument.Statement = append(policy.PolicyDocument.Statement, events.IAMPolicyStatement{
		Action:   []string{"execute-api:Invoke"},
		Effect:   "Allow",
		Resource: []string{event.MethodArn},
	})

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	roleSessionName := fmt.Sprintf("%s-%s", claims.TenantID, requestID)

	assumedRole, err := assumeRole(sess, roleArn, roleSessionName, iamPolicy)
	if err != nil {
		log.Printf("Request ID: %s Error assuming role: %v", requestID, err)
		return policy, err
	}

	policy.Context = map[string]interface{}{
		"accessKeyId":     *assumedRole.Credentials.AccessKeyId,
		"secretAccessKey": *assumedRole.Credentials.SecretAccessKey,
		"sessionToken":    *assumedRole.Credentials.SessionToken,
		"tenantId":        claims.TenantID,
		"userRole":        claims.UserRole,
		"email":           claims.Email,
		"region":          claims.Region,
		"awsRegion":       region,
		"firstName":       claims.FirstName,
		"lastName":        claims.LastName,
		"requestId":       requestID,
		"userId":          claims.Subject,
		// CORS headers
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "*",
		"Access-Control-Allow-Headers": "*",
		"Content-Type":                 "*/*",
	}

	tenantTier := "Premier" // TODO: Determine tnantTier from tenantDetails table

	usageIdentifierKey, err := getUsageIdentifierKey(tenantTier)

	if err != nil {
		log.Printf("Request ID: %s Error getting usage identifier key: %v", requestID, err)
	} else {
		log.Printf("Request ID: %s Usage identifier key: %s", requestID, usageIdentifierKey)
		policy.UsageIdentifierKey = usageIdentifierKey
	}

	// Marshal the context map to a JSON string with indentation
	_, err = json.MarshalIndent(policy.Context, "", "  ")
	if err != nil {
		log.Printf("Request ID: %s Error marshalling context: %v", requestID, err)
		return policy, err
	}

	log.Printf("Request ID: %s Accepted", requestID)

	return policy, nil

}

func main() {
	lambda.Start(Handler)
}
