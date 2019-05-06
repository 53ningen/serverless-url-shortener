package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/53ningen/serverless-url-shotener/models"
	"github.com/guregu/dynamo"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	// ErrInvalidURL invalid request parameter: url
	ErrInvalidURL = errors.New("invalid request parameter: url")
	// ErrInvalidTTL invalid request parameter: ttl
	ErrInvalidTTL = errors.New("invalid request parameter: ttl")
	// ErrInternal internal error
	ErrInternal = errors.New("internal error")

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

func main() {
	lambda.Start(handler)
}

func handler(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	givenURL, err := extractURL(event)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			Body:       "invalid request parameter",
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			Body:       getErrorMessage(ErrInternal.Error()),
			StatusCode: http.StatusServiceUnavailable,
		}, nil
	}

	result, err := handle(givenURL, &ttl)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			Body:       getErrorMessage(err.Error()),
			StatusCode: http.StatusServiceUnavailable,
		}, nil
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(result)
	if err != nil {
		log.Println(ErrInternal)
		return events.APIGatewayProxyResponse{
			Body:       getErrorMessage(err.Error()),
			StatusCode: http.StatusServiceUnavailable,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		Body:       string(buffer.Bytes()),
		StatusCode: http.StatusOK,
	}, nil
}

func extractURL(event events.APIGatewayProxyRequest) (givenURL *url.URL, err error) {
	var body models.CreateURLRequestBody
	e := json.Unmarshal([]byte(event.Body), &body)
	if e != nil {
		return nil, ErrInvalidURL
	}
	u, e := url.Parse(body.URL)
	if err != nil || !isValidURLString(u) || len(body.URL) > 1024 {
		return nil, ErrInvalidURL
	}
	return u, nil
}

func isValidURLString(u *url.URL) bool {
	return u.Scheme == "http" || u.Scheme == "https"
}

func getErrorMessage(msg string) string {
	m, _ := json.Marshal(map[string]string{
		"message": msg,
	})
	return string(m)
}

func handle(longURL *url.URL, ttl *int) (*models.URLMappingResult, error) {
	mapper := &models.URLMapper{
		HostName:     hostName,
		MappingStore: store,
	}
	return mapper.CreateMapping(longURL, time.Now(), ttl)
}
