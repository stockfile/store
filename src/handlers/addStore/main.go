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

	uuid "github.com/satori/go.uuid"
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

func AddStore(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("AddStore")

	var (
		id        = uuid.Must(uuid.NewV4(), nil).String()
		tableName = aws.String(os.Getenv("STORES_TABLE_NAME"))
		timestamp = time.Now().Format(time.RFC3339)
	)

	store := &Store{
		ID:        id,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}

	// Parse request body
	json.Unmarshal([]byte(request.Body), store)

	// Write to DynamoDB
	item, _ := dynamodbattribute.MarshalMap(store)
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: tableName,
	}

	if _, err := ddb.PutItem(input); err != nil {
		return events.APIGatewayProxyResponse{ // Error HTTP response
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	} else {
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
	lambda.Start(AddStore)
}
