# Utils Package

The `exec` package provides utilities for building an execution environment for Lambda functions within the Vantagea Control Plane. It extracts necessary contextual information from the incoming request, which can then be used throughout the execution of the Lambda.

## Usage

1. **Import the Package**

```go
import "github.com/tomweston/shared-service-authorizer/utils"
```

2. **Build Execution Environment**

Use the `BuildExecutionEnvironment` function to extract `AuthorizerContext` and `ExecutionContext` from the incoming `APIGatewayProxyRequest`. This function is typically called at the beginning of your Lambda handler.

```go
auth, exec, err := exec.BuildExecutionEnvironment(request)
if err != nil {
    // handle error
}
```

3. **Access Contextual Information**

The `AuthorizerContext` and `ExecutionContext` structs provide various pieces of contextual information extracted from the request, which can be used throughout your Lambda function.

### Example:

```go
fmt.Println(exec.AWSRegion)  // Output: us-west-2
fmt.Println(auth.AccessKeyID)  // Output: AKIAIOSFODNN7EXAMPLE
```

## Structs

- **AuthorizerContext**
  - `AccessKeyID`: AWS access key ID
  - `SecretAccessKey`: AWS secret access key
  - `SessionToken`: AWS session token
  - `ExecutionContext`: Embedded `ExecutionContext` struct

- **ExecutionContext**
  - `AWSRegion`: AWS region
  - `FirstName`: First name of the user
  - `LastName`: Last name of the user
  - `Email`: Email of the user
  - `TenantID`: Tenant ID
  - `UserRole`: User role

## Dependency

This package depends on:

- [github.com/aws/aws-lambda-go](https://github.com/aws/aws-lambda-go) for Lambda event structures
- [go.uber.org/zap](https://go.uber.org/zap) for logging

## Error Handling

If any of the expected information is missing from the request, `BuildExecutionEnvironment` will return an error. It's crucial to check and handle this error in your Lambda handler to ensure your function behaves correctly.
