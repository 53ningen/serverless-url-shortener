package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/53ningen/serverless-url-shotener/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

var (
	region    = os.Getenv("AWS_DEFAULT_REGION")
	ttlStr    = os.Getenv("TTL")
	hostName  = os.Getenv("HostName")
	tableName = os.Getenv("MappingTable")

	sess  = session.Must(session.NewSession())
	ddb   = dynamo.New(sess, aws.NewConfig().WithRegion(region))
	table = ddb.Table(tableName)

	cache = make(map[string]*models.URLMapping)
	store = models.DDBMappingStore{Cache: &cache, Table: &table}
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, ok := request.QueryStringParameters["id"]
	if !ok {
		return events.APIGatewayProxyResponse{
			Body:       getErrorMessage("invalid request parameter"),
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	m, e := store.GetURLMapping(id)
	if e != nil {
		log.Println(e.Error())
		return events.APIGatewayProxyResponse{
			Body:       getErrorMessage("internal server error"),
			StatusCode: http.StatusServiceUnavailable,
		}, nil
	}
	if m == nil {
		return events.APIGatewayProxyResponse{
			Body:       getErrorMessage("resource not found"),
			StatusCode: http.StatusNotFound,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusPermanentRedirect,
		Headers: map[string]string{
			"location": m.URL,
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}

func getErrorMessage(msg string) string {
	m, _ := json.Marshal(map[string]string{
		"message": msg,
	})
	return string(m)
}
