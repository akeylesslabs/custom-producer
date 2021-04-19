# Akeyless Custom Producer example

This is a dummy implementation of Akeyless Custom Producer that can be used as
an example of how to implement it and deploy using AWS Lambda.

## Installation

### Prerequisites

To use this custom producer, an AWS account is required.

### Setting up AWS Lambda

Please follow [this
guide](https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html) to
deploy this producer using AWS Lambda.

**TLDR**:

```sh
cd echoserver/bin/lambda
GOOS=linux CGO_ENABLED=0 go build -o main .
zip function.zip main
aws lambda create-function \
    --function-name my-function \
    --runtime go1.x \
    --zip-file fileb://function.zip \
    --handler main \
    --role arn:aws:iam::your-account-id:role/execution_role
```

### Creating AWS API Gateway

AWS Lambda function created in the previous step needs to be invoked using
HTTP. In AWS, this can be achieved using API Gateway. Please follow these steps
to expose this Akeyless Custom Producer implementation deployed as a Lambda
function:

1. In AWS console, navigate to API Gateway service.
1. If you already have at least one API, click "Create API" button.
1. Click "Build" under "HTTP API" section.
1. Under "Create and configure integrations", add an integration with AWS
   Lambda deployed during [Setting up AWS Lambda](#setting-up-aws-lambda) step,
   and select a name for the API, for example, "echoserver-producer".
1. Under "Configure routes" section, add two routes:
    ```
    POST /sync/create
    POST /sync/revoke
    ```
   Use the Lambda as "Integration target".
1. Keep "Define stages" step unchanged.
1. Confirm the settings and create the API.

### Setting up a producer

To setup the producer in Akeyless API Gateway, create a new "Custom Producer"
and use the URL created by API Gateway as the target endpoint:

```
https://some-id.execute-api.region.amazonaws.com/sync/create
https://some-id.execute-api.region.amazonaws.com/sync/revoke
```
