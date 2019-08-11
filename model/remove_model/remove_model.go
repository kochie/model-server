package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/kochie/model-server/model"
)

func deleteModelFromDynamo(sess *session.Session, modelName string, version string) error {
	tableName := os.Getenv("TABLE_NAME")
	svc := dynamodb.New(sess)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(modelName),
			},
			"version": {
				N: aws.String(version),
			},
		},
		TableName: aws.String(tableName),
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		return err
	}

	return nil
}

func deleteModelFromS3(sess *session.Session, modelName string, version string) error {
	svc := s3.New(sess)
	bucketName := os.Getenv("BUCKET_NAME")
	key := fmt.Sprintf("%s/%s.h5", modelName, version)
	request := s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	_, err := svc.DeleteObject(&request)
	if err != nil {
		return err
	}

	return nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	modelName, ok := request.PathParameters["model_name"]
	if !ok {
		errorString := "model_name was not specified in the url"
		log.Println(errorString)
		return model.CreateError(errorString, 400)
	}

	version, ok := request.QueryStringParameters["version"]
	if !ok {
		errorString := "version was not specified in the url parameter"
		log.Println(errorString)
		return model.CreateError(errorString, 400)
	}

	sess, err := session.NewSession()
	if err != nil {
		errorString := fmt.Sprintf("Could not make session %s", err.Error())
		fmt.Println(errorString)
		return model.CreateError(errorString, 500)
	}

	err = deleteModelFromDynamo(sess, modelName, version)
	if err != nil {
		errorString := err.Error()
		return model.CreateError(errorString, 500)
	}

	err = deleteModelFromS3(sess, modelName, version)
	if err != nil {
		errorString := err.Error()
		return model.CreateError(errorString, 500)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
