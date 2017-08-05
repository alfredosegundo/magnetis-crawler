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

	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var client *http.Client
var ctx context.Context

const FirstRow = 2

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

func UpdateEquityCurve(equities []magnetis.Equity, spreadsheetId string) (err error) {
	rowsCount := len(equities) + 1
	v := make([][]interface{}, rowsCount)
	v[0] = append(v[0], "Data", "Saldo Atual", "Total Aplicado", "Retorno", "Retorno dia", "Retorno dia %")
	for i, equity := range equities {
		currentSlicePos := i + 1
		currentRow := FirstRow + i
		v[currentSlicePos] = append(v[currentSlicePos],
			fmt.Sprintf("=DATE(%d,%d,%d)", equity.Time.Year(), equity.Time.Month(), equity.Time.Day()),
			fmt.Sprintf("=%s", equity.Value),
			fmt.Sprintf("=IF(B%d=\"\",\"\",SUMIF(Aplicado!$A$%d:A,\"<=\"&A%d,Aplicado!$B$%d:B))", currentRow, FirstRow, currentRow, FirstRow),
			fmt.Sprintf("=IF(B%d=\"\",\"\",B%d-C%d)", currentRow, currentRow, currentRow),
			fmt.Sprintf("=D%d-%s", currentRow, previousRow(currentRow)),
			fmt.Sprintf("=IF(B%d=0,0,E%d/B%d)", currentRow, currentRow, currentRow))
	}

	return updateSpreadSheet(v, spreadsheetId, fmt.Sprintf("Rendimento!A1:F%v", rowsCount))
}

func previousRow(currentRow int) (previousRow string) {
	if currentRow == FirstRow {
		return "0"
	}
	return fmt.Sprintf("D%d", currentRow-1)
}

func UpdateApplications(applications []magnetis.Application, spreadsheetId string) (err error) {
	rowsCount := len(applications) + 1
	v := make([][]interface{}, rowsCount)
	v[0] = append(v[0], "Data", "Tipo da transação", "Investimento", "Quantidade", "Preço (R$)", "IR (R$)", "Total Líquido (R$)")

	for i := range applications {
		application := applications[i]
		v[i+1] = append(v[i+1],
			fmt.Sprintf("=DATE(%d,%d,%d)", application.Date.Year(), application.Date.Month(), application.Date.Day()),
			application.Type.String(),
			strings.TrimSpace(application.Investment),
			fmt.Sprintf("=%f", application.Quantity),
			fmt.Sprintf("=%f", application.Price),
			fmt.Sprintf("=%f", application.IR),
			fmt.Sprintf("=%f", application.Net),
		)
	}
	return updateSpreadSheet(v, spreadsheetId, fmt.Sprintf("Historico!A1:G%v", rowsCount))
}

func updateSpreadSheet(values [][]interface{}, spreadsheetId string, valuesRange string) (err error) {
	service, err := sheets.New(client)
	if err != nil {
		return err
	}
	rb := &sheets.ValueRange{Values: values, MajorDimension: "ROWS"}
	valueInputOption := "USER_ENTERED"
	_, err = service.Spreadsheets.Values.Update(spreadsheetId, valuesRange, rb).ValueInputOption(valueInputOption).Context(ctx).Do()
	if err != nil {
		return err
	}
	return
}
