<div align="center">

# shared-service-authorizer

[Key Features](#key-features) •
[Prerequisites](#prerequisites) •
[Deployment](#-deployment) •
[Contributing](#-contributing) •
[License](#-license)

The Shared Service Authorizer is inspired by the AWS SaaS Factory Serverless SaaS Identity and Isolation Patterns Lambda function invoked by API Gateway to authorize requests.

</div>

## Overview

- Validates JWT tokens.
- Generates STS credentials and an IAM policy for tenants.
- Returns the policy and credentials to API Gateway.

## Key Features

- User role management with distinct permissions for System Admins, Tenant Admins, and Users.
- Dynamic IAM policy generation based on user roles and tenant IDs.
- Integration with AWS services like DynamoDB and STS for access control and session management.
- Customizable for different regions and service identifiers.

[Saas Identity and Isolation Patterns](./docs/SaaS_tenant_isolation_patterns.pdf)

## Serverless Configuration

The `serverless.yml` file defines the AWS services and resources required for the `shared-service-authorizer` service. Key components of this file are:

### Functions

- **Authorizer**: A Lambda function written in Go that serves as the authorizer for the API Gateway. It validates incoming requests based on the provided JWT tokens.

- **CreateHelloItem**: Another Lambda function in Go. This function is triggered via an HTTP POST request to `/v1/hello` and is responsible for creating items in a DynamoDB table.

### Resources

- **ApiGatewayRestApi**: Defines the API Gateway REST API for the shared services.

- **ApiGatewayAuthorizer**: Configures a custom authorizer for the API Gateway. It specifies the authorizer URI, identity source, and TTL settings.

- **LambdaApiGatewayInvoke**: Grants the API Gateway permission to invoke the Authorizer function.

- **AuthorizerLambdaRole and CreateHelloItemLambdaRole**: IAM roles for the respective Lambda functions, granting them necessary permissions like interacting with DynamoDB, logging, and assuming other roles.

- **SharedServices DynamoDB Table**: Defines the DynamoDB table used by the application. It follows a Single Table Design with specified attribute definitions, key schema, and global secondary indexes.

### Plugins

- **serverless-go-plugin**: Facilitates the building of Go-based Lambda functions.

- **serverless-plugin-datadog**: Integrates Datadog monitoring with the Serverless setup.

### Custom

- Configuration for Datadog integration, including API key setup.

### Outputs

- References to resources like the DynamoDB table ARN and IAM roles, which can be used elsewhere in the serverless setup.

This configuration file is essential for deploying and managing your Lambda functions and related AWS resources using the Serverless Framework.

</div>

## Prerequisites

- Ensure you have `Go` installed on your machine.
- Set up and configure your AWS credentials for Pulumi deployments.

## 🚀 Deployment

The application and its infrastructure are managed and deployed using Serverless Framework.

- Make sure you have the Serverless CLI installed. If not, install it using:

    ```bash
    npm install -g serverless
    ```

### Deployment Steps

1. Deploy the service using the following command:

    ```bash
    sls deploy --stage dev
    ```

2. **Deployment:** The Serverless Framework is used to deploy the application to the specified stage.

    ```yaml
    - name: Deploy Service
      run: sls deploy --stage dev
    ```

3. **Test:** The Serverless Framework is used to invoke the Lambda function and pass in a test event.

    ```yaml
    - name: Test Service
      run: sls invoke --stage dev --function authoriser --path test/event.json
    ```

4. **Cleanup:** The Serverless Framework is used to remove the application from AWS.

    ```yaml
    - name: Remove Service
      run: sls remove --stage dev
    ```

## 🤝 Contributing

Contributions, issues and feature requests are welcome!

## 📄 License

This project is [MIT](./LICENSE) licensed.

## 👨‍💻 Author

<h2  align="center">📦 Reach me on</h2>
<p align="center">
  <a target="_blank"href="https://www.linkedin.com/in/westontom"><img src="https://img.shields.io/badge/linkedin-%230077B5.svg?&style=for-the-badge&logo=linkedin&logoColor=white" /></a>&nbsp;&nbsp;&nbsp;&nbsp;
  <a target="_blank"href="https://twitter.com/tomweston"><img src="https://img.shields.io/badge/@tomweston-%231DA1F2.svg?&style=for-the-badge&logo=x&logoColor=white" /></a>&nbsp;&nbsp;&nbsp;&nbsp;
  <a href="mailto:weston.tom@gmail.com?subject=Hello%20Tom,%20From%20Github"><img src="https://img.shields.io/badge/gmail-%23D14836.svg?&style=for-the-badge&logo=gmail&logoColor=white" /></a>&nbsp;&nbsp;&nbsp;&nbsp;
</p>