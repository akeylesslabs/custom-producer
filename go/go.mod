module github.com/akeylesslabs/custom-producer/go

go 1.16

require (
	github.com/aws/aws-lambda-go v1.23.0
	github.com/go-acme/lego/v4 v4.3.1
	github.com/gorilla/mux v1.8.0
)

exclude github.com/labstack/echo/v4 v4.1.11
