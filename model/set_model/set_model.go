package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/kochie/model-server/model"
)

func setModelInDynamo(sess *session.Session, machineModel *model.Model) error {
	tableName := os.Getenv("TABLE_NAME")
	svc := dynamodb.New(sess)
	av, err := dynamodbattribute.MarshalMap(machineModel)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return err
	}
	return nil
}

func setModelInS3(sess *session.Session, modelName string, version int) (string, error) {
	svc := s3.New(sess)
	bucketName := os.Getenv("BUCKET_NAME")
	key := fmt.Sprintf("%s/%d.h5", modelName, version)

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(15 * time.Minute)

	if err != nil {
		errorString := fmt.Sprintf("Failed to sign request because %s", err)
		log.Println(errorString)
		return "", errors.New(errorString)
	}
	return urlStr, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	machineModel := model.Model{}
	err := json.Unmarshal([]byte(request.Body), &machineModel)
	if err != nil {
		errorString := fmt.Sprintf("Could not parse json %s", err.Error())
		fmt.Println(errorString)
		return model.CreateError(errorString, 400)
	}

	sess, err := session.NewSession()
	if err != nil {
		errorString := fmt.Sprintf("Could not make session %s", err.Error())
		fmt.Println(errorString)
		return model.CreateError(errorString, 500)
	}

	err = setModelInDynamo(sess, &machineModel)
	if err != nil {
		errorString := err.Error()
		return model.CreateError(errorString, 500)
	}

	urlStr, err := setModelInS3(sess, machineModel.Name, machineModel.Version)
	if err != nil {
		errorString := err.Error()
		return model.CreateError(errorString, 500)
	}

	body, err := json.Marshal(struct {
		*model.Model
		URL string `json:"uploadUrl"`
	}{&machineModel, urlStr})
	if err != nil {
		errorString := fmt.Sprintf("Error marshaling JSON: %s", err.Error())
		return model.CreateError(errorString, 500)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
