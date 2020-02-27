package spreadsheet

import (
	"fmt"
	"io/ioutil"
	"log"

	"net/http"

	"net/url"
	"os"
	"path/filepath"

	"encoding/json"

	"os/user"

	"strings"

	"github.com/alfredosegundo/magnetis-crawler/magnetis"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var client *http.Client
var ctx context.Context

const firstRow = 2

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
	var f *os.File
	var err error
	t := &oauth2.Token{}
	if f, err = os.Open(file); err != nil {
		if envvar, exist := os.LookupEnv("CREDENTIALS"); exist {
			err = json.NewDecoder(strings.NewReader(envvar)).Decode(t)
			return t, nil
		}
		return t, err
	}
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
		log.Println("token cache file not found.")
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func SpreadsheetsSignin() {
	ctx := context.Background()
	var b []byte
	var err error
	if b, err = ioutil.ReadFile("client_secret.json"); err != nil {
		if envvar, exist := os.LookupEnv("CLIENT_SECRET"); exist {
			b = []byte(envvar)
		} else {
			log.Fatalf("Unable to read client secret env var. It's not set.")
		}
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/sheets.googleapis.com-go-quickstart.json
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client = getClient(ctx, config)
}

func sumAsset(firstRow int, currentRow int, assetName magnetis.TransactionType) (formula string) {
	return fmt.Sprintf("SUMIFS(Historico!$H$%v:$H,Historico!$A$%v:$A,\"<=\"&$A%v,Historico!$C$%v:$C,\"=%v\")", firstRow, firstRow, currentRow, firstRow, assetName)
}

func UpdateEquityCurve(equities []magnetis.Equity, spreadsheetID string) (err error) {
	rowsCount := len(equities) + 1
	v := make([][]interface{}, rowsCount)
	v[0] = append(v[0], "Data", "Saldo Atual", "Total Aplicado", "Retorno", "Retorno dia", "Retorno dia %",
		"Retorno desde início", "R$/R$ investido", "Mês", "Ano", "Total Aplicado")
	for i, equity := range equities {
		currentSlicePos := i + 1
		currentRow := firstRow + i
		v[currentSlicePos] = append(v[currentSlicePos],
			fmt.Sprintf("=DATE(%d,%d,%d)", equity.Time.Year(), equity.Time.Month(), equity.Time.Day()),
			fmt.Sprintf("=%s", equity.Value),
			fmt.Sprintf("=SUMIF(Aplicado!$A$%d:A,\"<=\"&A%d,Aplicado!$B$%d:B)", firstRow, currentRow, firstRow),
			fmt.Sprintf("=B%d-C%d", currentRow, currentRow),
			fmt.Sprintf("=D%d-%s", currentRow, previousRow(currentRow)),
			fmt.Sprintf("=E%d/B%d", currentRow, currentRow),
			fmt.Sprintf("=SUM($F$%d:F%d)", firstRow, currentRow),
			fmt.Sprintf("=D%d/C%d", currentRow, currentRow),
			fmt.Sprintf("=%d", equity.Time.Month()),
			fmt.Sprintf("=%d", equity.Time.Year()),
			fmt.Sprintf("=%v-%v-%v+%v+%v",
				sumAsset(firstRow, currentRow, magnetis.MoneyApplication),
				sumAsset(firstRow, currentRow, magnetis.Redemption),
				sumAsset(firstRow, currentRow, magnetis.ExpiredTitle),
				sumAsset(firstRow, currentRow, magnetis.AdvisoryFee),
				sumAsset(firstRow, currentRow, magnetis.TransactionFees)))
	}

	return updateSpreadSheet(v, spreadsheetID, fmt.Sprintf("Rendimento!A1:K%v", rowsCount))
}

func previousRow(currentRow int) (previousRow string) {
	if currentRow == firstRow {
		return "0"
	}
	return fmt.Sprintf("D%d", currentRow-1)
}

func UpdateApplications(applications []magnetis.Application, spreadsheetID string) (err error) {
	rowsCount := len(applications) + 1
	v := make([][]interface{}, rowsCount)
	v[0] = append(v[0], "Data aplicação", "Data efetivação", "Tipo da transação", "Investimento", "Quantidade", "Preço (R$)", "IR (R$)", "Total Líquido (R$)")

	for i := range applications {
		application := applications[i]
		v[i+1] = append(v[i+1],
			fmt.Sprintf("=DATE(%d,%d,%d)", application.ApplicationDate.Year(), application.ApplicationDate.Month(), application.ApplicationDate.Day()),
			fmt.Sprintf("=DATE(%d,%d,%d)", application.Date.Year(), application.Date.Month(), application.Date.Day()),
			application.Type.String(),
			strings.TrimSpace(application.Investment),
			fmt.Sprintf("=%f", application.Quantity),
			fmt.Sprintf("=%f", application.Price),
			fmt.Sprintf("=%f", application.IR),
			fmt.Sprintf("=%f", application.Net),
		)
	}
	return updateSpreadSheet(v, spreadsheetID, fmt.Sprintf("Historico!A1:H%v", rowsCount))
}

func updateSpreadSheet(values [][]interface{}, spreadsheetID string, valuesRange string) (err error) {
	service, err := sheets.New(client)
	if err != nil {
		return err
	}
	rb := &sheets.ValueRange{Values: values, MajorDimension: "ROWS"}
	valueInputOption := "USER_ENTERED"
	_, err = service.Spreadsheets.Values.Update(spreadsheetID, valuesRange, rb).ValueInputOption(valueInputOption).Context(ctx).Do()
	if err != nil {
		return err
	}
	return
}

func GetConfiguredStocks() (stocks []string) {
	return []string{"VVAR3", "MGLU3"}
}

func UpdateStocks(stock string, value string) {
	fmt.Printf("%s: %s\n", stock, value)
}
