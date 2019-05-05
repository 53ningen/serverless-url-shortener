package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandler(t *testing.T) {
	t.Run("extractURL", func(t *testing.T) {
		queryStringParameters := make(map[string]string)
		queryStringParameters["url"] = "https://example.com/path/to/resources?param1=value1&param2=value2"
		req := events.APIGatewayProxyRequest{
			QueryStringParameters: queryStringParameters,
		}
		u, e := extractURL(req)
		if e != nil {
			t.Fatalf("failure extractURL:  %s", e)
		}
		actual := (*u).String()
		expected := "https://example.com/path/to/resources?param1=value1&param2=value2"
		if actual != expected {
			t.Fatalf("getMD5Hash failure: expected %s, actual %s", expected, actual)
		}
	})
}
