package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alfredosegundo/magnetis-crawler/magnetis"

	"github.com/aws/aws-lambda-go/lambda"
)

// MyEvent is the event dispatched by aws lambda infrastrucute
// to start the function
type MyEvent struct {
	Name string `json:"name"`
}

// HandleRequest is the entrypoint of the lambda function
func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	err := magnetis.MagnetisSignin(os.Getenv("MAGNETIS_USER_ID"), os.Getenv("MAGNETIS_PASSWORD"))
	if err != nil {
		return fmt.Sprintf("Failed to login: \n%v", err), err
	}
	plan, err := magnetis.GetInvestmentPlan(os.Getenv("USER_ID"))
	if err != nil {
		return fmt.Sprintf("Failed to get investment plan: \n%v", err), err
	}

	return fmt.Sprintf("Plan: %v", plan), nil
}

func main() {
	lambda.Start(HandleRequest)
}
