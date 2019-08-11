package model

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

// CreateError will return an error response for lambda function.
func CreateError(errorString string, statusCode int) (events.APIGatewayProxyResponse, error) {
	body, err := json.Marshal(struct {
		Error string
	}{errorString})

	if err != nil {
		log.Println("Had an error creating an error, fuck did you write")
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: statusCode,
	}, nil
}
