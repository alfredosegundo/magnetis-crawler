// Package spreadsheet provides an client to interact with google drive spreadsheets
package spreadsheet

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/alfredosegundo/magnetis-crawler/magnetis"

	"net/http"

	"net/url"
	"os"
	"path/filepath"

	"encoding/json"

	"os/user"

	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var client *http.Client
var ctx context.Context

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".magnetis_crawler")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("credentials.json")), nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func Signin() {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/sheets.googleapis.com-go-quickstart.json
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client = getClient(ctx, config)
}

func UpdateEquityCurve(curve *magnetis.EquityCurve) (err error) {

	equities := curve.Equities
	rowData := make([]*sheets.RowData, len(equities))
	rowData = append(rowData, &sheets.RowData{Values: []*sheets.CellData{
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: "Data"}},
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: "Saldo Atual"}}}})
	for i := range equities {
		if err != nil {
			log.Println(err)
		}
		equity := equities[i]
		rowData = append(rowData, &sheets.RowData{Values: []*sheets.CellData{
			createDateCell(equity.Time.Year(), equity.Time.Month(), equity.Time.Day()),
			createStringMoneyCell(equity.Value)}})
	}
	equitySheet := &sheets.Sheet{Data: []*sheets.GridData{{RowData: rowData}}, Properties: &sheets.SheetProperties{Title: "EquityCurve"}}
	rb := &sheets.Spreadsheet{
		Sheets:     []*sheets.Sheet{equitySheet},
		Properties: &sheets.SpreadsheetProperties{Title: "Planejamento"},
	}
	service, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}
	_, err = service.Spreadsheets.Create(rb).Context(ctx).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	return
}

func UpdateApplications(applications []magnetis.Application) (err error) {
	service, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}
	rowData := make([]*sheets.RowData, len(applications))
	rowData = append(rowData, &sheets.RowData{Values: []*sheets.CellData{
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: "Data"}},
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: "Tipo da transação"}},
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: "Investimento"}},
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: "Quantidade"}},
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: "Preço (R$)"}},
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: "IR (R$)"}},
		{UserEnteredValue: &sheets.ExtendedValue{StringValue: "Total Líquido (R$)"}},
	}})
	for i := range applications {
		application := applications[i]
		rowData = append(rowData, &sheets.RowData{Values: []*sheets.CellData{
			createDateCell(application.Date.Year(), application.Date.Month(), application.Date.Day()),
			createStringCell(application.Type.String()),
			createStringCell(application.Investment),
			createFloatMoneyCell(application.Quantity),
			createFloatMoneyCell(application.Price),
			createFloatMoneyCell(application.IR),
			createFloatMoneyCell(application.Net)}})
	}
	equitySheet := &sheets.Sheet{Data: []*sheets.GridData{{RowData: rowData}}, Properties: &sheets.SheetProperties{Title: "History"}}
	rb := &sheets.Spreadsheet{
		Sheets:     []*sheets.Sheet{equitySheet},
		Properties: &sheets.SpreadsheetProperties{Title: "Planejamento"},
	}
	_, err = service.Spreadsheets.Create(rb).Context(ctx).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	return
}

func createDateCell(year int, month time.Month, day int) *sheets.CellData {
	return &sheets.CellData{
		UserEnteredFormat: &sheets.CellFormat{
			NumberFormat: &sheets.NumberFormat{
				Type: "DATE", Pattern: "ddd\", \"d\"/\"m\"/\"yy"}},
		UserEnteredValue: &sheets.ExtendedValue{
			FormulaValue: fmt.Sprintf("=DATE(%d,%d,%d)", year, month, day)}}
}

func createStringCell(stringValue string) *sheets.CellData {
	return &sheets.CellData{
		UserEnteredValue: &sheets.ExtendedValue{
			StringValue: stringValue},
	}
}

func createFloatMoneyCell(value float64) *sheets.CellData {
	return createStringMoneyCell(fmt.Sprintf("%f", value))
}

func createStringMoneyCell(value string) *sheets.CellData {
	return &sheets.CellData{
		UserEnteredFormat: &sheets.CellFormat{
			NumberFormat: &sheets.NumberFormat{Type: "CURRENCY"}},
		UserEnteredValue: &sheets.ExtendedValue{
			FormulaValue: fmt.Sprintf("=%s", value)}}
}
