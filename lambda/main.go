package main

import (
	"context"
	"log"
	"os"

	"github.com/alfredosegundo/magnetis-crawler/magnetis"
	"github.com/alfredosegundo/magnetis-crawler/spreadsheet"

	"github.com/aws/aws-lambda-go/lambda"
)

// MyEvent is the event dispatched by aws lambda infrastrucute
// to start the function
type MyEvent struct {
	Name string `json:"name"`
}

// HandleRequest is the entrypoint of the lambda function
func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	magnetisUserID := os.Getenv("MAGNETIS_USER")
	magnetisPassword := os.Getenv("MAGNETIS_PASS")
	userID := os.Getenv("MAGNETIS_USER_ID")
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	err := magnetis.MagnetisSignin(magnetisUserID, magnetisPassword)

	if err != nil {
		log.Fatal(err)
	}
	spreadsheet.SpreadsheetsSignin()
	curve, err := magnetis.GetEquityCurve(userID)
	if err != nil {
		log.Fatal(err)
	}

	err = spreadsheet.UpdateEquityCurve(curve.Equities, spreadsheetID)
	if err != nil {
		log.Fatal(err)
	}

	return "done", nil
}

func main() {
	lambda.Start(HandleRequest)
}
