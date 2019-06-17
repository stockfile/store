package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Response struct {
	Store Store `json:"store"`
}

type Store struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

var ddb *dynamodb.DynamoDB

func init() {
	region := os.Getenv("AWS_REGION")
	if session, err := session.NewSession(&aws.Config{
		Region: &region,
	}); err != nil {
		fmt.Println(fmt.Sprintf("Failed to connect to AWS: %s", err.Error()))
	} else {
		ddb = dynamodb.New(session) // Create DynamoDB client
	}
}

func UpdateStore(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("UpdateStore")

	var (
		id        = request.PathParameters["id"]
		tableName = aws.String(os.Getenv("STORES_TABLE_NAME"))
		timestamp = time.Now().Format(time.RFC3339)
	)

	store := &Store{
		UpdatedAt: timestamp,
	}

	// Parse request body
	json.Unmarshal([]byte(request.Body), store)

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#d":    aws.String("description"),
			"#u_at": aws.String("updated_at"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":d": {
				S: aws.String(store.Description),
			},
			":u_at": {
				S: aws.String(store.UpdatedAt),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String("SET #d = :d, #u_at = :u_at"),
		ReturnValues:     aws.String("ALL_NEW"),
		TableName:        tableName,
	}

	if result, err := ddb.UpdateItem(input); err != nil {
		return events.APIGatewayProxyResponse{ // Error HTTP response
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	} else {
		store := new(Store)
		err = dynamodbattribute.UnmarshalMap(result.Attributes, store)

		if err != nil {
			return events.APIGatewayProxyResponse{ // Error HTTP response
				Body:       err.Error(),
				StatusCode: 500,
			}, nil
		}

		body, _ := json.Marshal(&Response{
			Store: *store,
		})

		return events.APIGatewayProxyResponse{ // Success HTTP response
			Body:       string(body),
			StatusCode: 200,
		}, nil
	}
}

func main() {
	lambda.Start(UpdateStore)
}
