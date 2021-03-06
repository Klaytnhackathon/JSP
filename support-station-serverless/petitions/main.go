package main

import (
	"petitions/db"
	"petitions/models"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		if req.PathParameters["petitionID"] == "" {
			return index(req)
		} else {
			return show(req)
		}
	case "POST":
		return create(req)
	case "PATCH":
		return update(req)
	default:
		return clientError(http.StatusMethodNotAllowed)
	}
}

func index(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	petitions, err := db.GetItems(&req.QueryStringParameters)
	if err != nil {
		return serverError(err)
	}

	totalCount, err := db.GetTotalCount()

	petitionsObj := map[string]interface{}{}
	petitionsObj["petitions"] = petitions
	petitionsObj["meta"] = models.Meta{TotalCount: totalCount}

	finalized, err := json.Marshal(petitionsObj)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers(),
		Body:       string(finalized),
	}, nil
}

func show(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ID := req.PathParameters["petitionID"]

	pt, err := db.GetItem(ID)
	if err != nil {
		return serverError(err)
	}
	if pt == nil {
		return clientError(http.StatusNotFound)
	}

	js, err := json.Marshal(pt)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers(),
		Body:       string(js),
	}, nil
}

func create(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	pt := new(models.Petition)
	err := json.Unmarshal([]byte(req.Body), pt)
	if err != nil {
		errorLogger.Fatalln(err)
		return clientError(http.StatusUnprocessableEntity)
	}

	petition, dbErr := db.PutItem(pt)
	if dbErr != nil {
		return serverError(dbErr)
	}

	petitionJSON, err := json.Marshal(petition)
	if err != nil {
		return serverError(err)
	}
	
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers:    headers(),
		Body: string(petitionJSON),
	}, nil
}

func update(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	pt := new(models.Petition)
	err := json.Unmarshal([]byte(req.Body), pt)
	if err != nil {
		errorLogger.Fatalln(err)
		return clientError(http.StatusUnprocessableEntity)
	}

	if err != nil {
		errorLogger.Fatalln(err)
		return clientError(http.StatusNotFound)
	}

	dbErr := db.UpdateItem(pt)
	if dbErr != nil {
		return serverError(dbErr)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 204,
		Headers:    headers(),
	}, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Headers:    headers(),
		Body:       err.Error(),
	}, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers:    headers(),
		Body:       http.StatusText(status),
	}, nil
}

func headers() map[string]string {
	return map[string]string{"Access-Control-Allow-Origin": "*", "Access-Control-Allow-Credentials": "true"}
}

func main() {
	lambda.Start(router)
}
