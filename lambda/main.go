package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alfredosegundo/magnetis-crawler/magnetis"

	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	err := magnetis.MagnetisSignin(os.Getenv("MAGNETIS_USER_ID"), os.Getenv("MAGNETIS_PASSWORD"))
	if err != nil {
		log.Fatal(err)
	}
	plan, err := magnetis.GetInvestmentPlan(os.Getenv("USER_ID"))
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("Plan: %v", plan), nil
}

func main() {
	lambda.Start(HandleRequest)
}
