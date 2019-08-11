package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kochie/model-server/model"
)

func getModelFromDynamo(sess *session.Session, modelName string, version string) (*model.Model, error) {
	tableName := os.Getenv("TABLE_NAME")
	svc := dynamodb.New(sess)

	key := map[string]*dynamodb.Condition{
		"name": {
			ComparisonOperator: aws.String("EQ"),
			AttributeValueList: []*dynamodb.AttributeValue{{
				S: aws.String(modelName),
			}},
		},
	}
	if version != "" {
		key["version"] = &dynamodb.Condition{
			ComparisonOperator: aws.String("EQ"),
			AttributeValueList: []*dynamodb.AttributeValue{{
				N: aws.String(version),
			}},
		}
	}

	result, err := svc.Query(&dynamodb.QueryInput{
		TableName:        aws.String(tableName),
		KeyConditions:    key,
		ScanIndexForward: aws.Bool(false),
	})

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	machineModel := model.Model{}

	err = dynamodbattribute.UnmarshalMap(result.Items[0], &machineModel)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if machineModel.Name == "" {
		errorString := fmt.Sprintf("Could not find %s", modelName)
		fmt.Println(errorString)
		return nil, errors.New(errorString)
	}

	return &machineModel, nil
}

func getSignedS3Link(sess *session.Session, bucketName string, key string) (string, error) {
	svcS3 := s3.New(sess)

	req, _ := svcS3.GetObjectRequest(&s3.GetObjectInput{
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
	modelName, ok := request.PathParameters["model_name"]
	if !ok {
		log.Println("model_name was not specified in the url")
		return model.CreateError("model_name was not specified in the url", 400)
	}

	version, ok := request.QueryStringParameters["version"]

	sess, err := session.NewSession()
	if err != nil {
		errorString := fmt.Sprintf("Could not make session %s", err.Error())
		fmt.Println(errorString)
		return model.CreateError(errorString, 400)
	}

	machineModel, err := getModelFromDynamo(sess, modelName, version)
	if err != nil {
		errorString := fmt.Sprintf("Could not get model from dynamo %s", err.Error())
		fmt.Println(errorString)
		return model.CreateError(errorString, 400)
	}

	urlStr, err := getSignedS3Link(
		sess, os.Getenv("BUCKET_NAME"), fmt.Sprintf("%s/%d.h5", machineModel.Name, machineModel.Version))
	if err != nil {
		fmt.Println(err.Error())
		return model.CreateError(err.Error(), 500)
	}

	body, err := json.Marshal(struct {
		*model.Model
		URL string `json:"downloadUrl"`
	}{machineModel, urlStr})

	if err != nil {
		log.Println("Failed to marshal json response", err)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
