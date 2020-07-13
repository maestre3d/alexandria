package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/blob-handler-service/serverless/viewer-request/pkg"
	"log"
)

var (
	muxLambda *gorillamux.GorillaMuxAdapter
)

func init() {
	// Stdout logging gets extracted by AWS CloudWatch automatically
	log.Printf("http routing start")
	r := mux.NewRouter()

	// Map routes with image-containing S3 endpoints
	r.Get("/alexandria/user/{id}").HandlerFunc(pkg.GetContentHandler)
	r.Get("/alexandria/user/cover/{id}").HandlerFunc(pkg.GetContentHandler)

	r.Get("/alexandria/author/{id}").HandlerFunc(pkg.GetContentHandler)
	r.Get("/alexandria/author/cover/{id}").HandlerFunc(pkg.GetContentHandler)

	r.Get("/alexandria/media/cover/{id}").HandlerFunc(pkg.GetContentHandler)
	r.Get("/alexandria/media/canvas/{id}").HandlerFunc(pkg.GetContentHandler)

	r.Get("/alexandria/ad/{id}").HandlerFunc(pkg.GetContentHandler)

	muxLambda = gorillamux.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return muxLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
