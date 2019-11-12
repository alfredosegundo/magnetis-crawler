package magnetis

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"sort"
	"time"

	"net/url"

	"strings"

	"strconv"

	"fmt"

	"github.com/PuerkitoBio/goquery"
)

var host = "https://magnetis.com.br"

// An Equity represents the amount of money if all of the assets were liquidated.
type Equity struct {
	Time  time.Time // Day when the value was measured
	Value string    // Amount of money
}

// Excel prints an equity as two excel cells separated by tabs \t
func (e Equity) Excel() string {
	et := e.Time
	return fmt.Sprintf("=DATE(%d,%d,%d)\t=%s", et.Year(), et.Month(), et.Day(), e.Value)
}

func (e Equity) String() string {
	return fmt.Sprintf("%v\t%s", e.Time, e.Value)
}

// An EquityCurve holds a slice of Equity
type EquityCurve struct {
	Equities []Equity
}

func (e EquityCurve) Len() int      { return len(e.Equities) }
func (e EquityCurve) Swap(i, j int) { e.Equities[i], e.Equities[j] = e.Equities[j], e.Equities[i] }
func (e EquityCurve) Less(i, j int) bool {
	return e.Equities[i].Time.Before(e.Equities[j].Time)
}

// An InvestmentPlan holds the original plan for the magnetis account plan
type InvestmentPlan struct {
	Age               int
	GoalValue         float64 `json:"goal_value,string"`
	InitialInvestment float32 `json:"initial_investment,string"`
	MonthlyInvestment float32 `json:"monthly_investment,string"`
	PeriodInYears     int     `json:"period_in_years"`
	RiskLevel         int     `json:"risk_level"`
}

// An Asset is an investment acquired for the account
type Asset struct {
	Amount             string
	AssetID            int    `json:"asset_id"`
	AssetReturn        string `json:"asset_return"`
	CategoryKey        string `json:"category_key"`
	InstrumentTypeName string `json:"instrument_type_name"`
	Issuer             string
	Liquidity          int
	MaturityDate       string `json:"maturity_date"`
	Name               string
	Yield              string
}

// TransactionType represents which transactions was performed with an Asset
type TransactionType int

// TransactionType codes for each type
const (
	MoneyApplication TransactionType = iota
	IRWithdrawal
	TransactionFees
	AdvisoryFee
	Redemption
	ExpiredTitle
)

var transactionTypes = [...]string{
	"Application",
	"IRWithdrawal",
	"TransactionFees",
	"AdvisoryFee",
	"Redemption",
	"Expired",
}

func (t TransactionType) String() string { return transactionTypes[t] }

type Application struct {
	Date            time.Time
	ApplicationDate time.Time
	Type            TransactionType
	Investment      string
	Quantity        float64
	Price           float64
	IR              float64
	Net             float64
}

func (a Application) String() string {
	return fmt.Sprintf("%v\t%v\t%s\t%s\t%f\t%f\t%f\t%f", a.ApplicationDate, a.Date, a.Investment, a.Type, a.Quantity, a.Price, a.IR, a.Net)
}

func (a Application) Excel() string {
	return fmt.Sprintf("=DATE(%d,%d,%d)\t%s\t%s\t=%f\t=%f\t=%f\t=%f", a.Date.Year(), a.Date.Month(), a.Date.Day(), a.Investment, a.Type, a.Quantity, a.Price, a.IR, a.Net)
}

var jar, _ = cookiejar.New(nil)
var defaultClient = &http.Client{
	Jar: jar,
}

func GetEquityCurve(userId string) (curve *EquityCurve, err error) {
	uri := host + "/pricing/api/portfolio/" + userId + "/equity_curve"
	log.Println(fmt.Sprintf("Equity curve url: %s", uri))
	resp, err := defaultClient.Get(uri)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d\nbody: %s", resp.StatusCode, string(body))
	}

	icurve := make([][]interface{}, 0)

	err = json.Unmarshal(body, &icurve)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal response body: %s", string(body))
	}
	curve = new(EquityCurve)
	for i := range icurve {
		equityTime := int64(icurve[i][0].(float64)) / 1000
		equity := Equity{Time: time.Unix(equityTime, 0).UTC(), Value: icurve[i][1].(string)}
		curve.Equities = append(curve.Equities, equity)
	}
	sort.Sort(curve)
	return
}

func MagnetisSignin(username string, password string) (err error) {
	var signin = host + "/users/sign_in"
	log.Printf("singing in on: %s", signin)
	resp, err := defaultClient.Get(signin)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("ERROR\nhttp status code: %d\nbody: %s", resp.StatusCode, string(body))
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return err
	}
	selection := doc.Find("input[name='authenticity_token']")
	token, _ := selection.First().Attr("value")
	_, err = defaultClient.PostForm(signin, url.Values{
		"authenticity_token": {token},
		"utf8":               {"✓"},
		"user[email]":        {username},
		"user[password]":     {password},
	})
	return err
}

func GetInvestmentPlan(userId string) (plan *InvestmentPlan, err error) {
	resp, err := defaultClient.Get(host + "/api/investment_plan/" + userId)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &plan)
	if err != nil {
		return nil, err
	}
	return
}

func Assets(userId string) (assets []Asset, err error) {
	resp, err := defaultClient.Get(host + "/user_portfolio/api/portfolios/" + userId + "/assets")
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &assets)
	if err != nil {
		return nil, err
	}
	return
}

func Applications() (applications []Application, err error) {
	res, err := defaultClient.Get(host + "/movimentacoes")
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}
	rows := doc.Find("section.transactions table tbody tr")
	for i := range rows.Nodes {
		investmentRow := rows.Eq(i)
		anTransaction := Application{
			Investment: investmentRow.Find("td:nth-child(2)").Text(),
			Quantity:   convertPtToEnNumber(investmentRow.Find("td:nth-child(3)").Text()),
			Price:      convertPtToEnNumber(investmentRow.Find("td:nth-child(4)").Text()),
			IR:         convertPtToEnNumber(investmentRow.Find("td:nth-child(5)").Text()),
			Net:        convertPtToEnNumber(investmentRow.Find("td:nth-child(6)").Text()),
		}
		if val, exists := investmentRow.Find("time").Attr("datetime"); exists {
			investmentDate, err := time.Parse("2006-01-02", val)
			if err != nil {
				return nil, err
			}
			anTransaction.Date = investmentDate
		} else {
			if !investmentRow.HasClass("advisory-fee") {
				anTransaction.Date = applications[len(applications)-1].Date
			}
		}
		dateElement := investmentRow.ParentsFiltered("div.user-order__header").First().Find("header time")
		if applicationDate, exists := dateElement.Attr("datetime"); exists {
			if anTransaction.ApplicationDate, err = time.Parse("2006-01-02", applicationDate); err != nil {
				break
			}
		}
		if investmentRow.HasClass("journal-summary__transaction-fees") {
			anTransaction.Type = TransactionFees
		}
		if investmentRow.HasClass("journal-summary__asset-trade--with-transaction-fees") {
			anTransaction.Type = MoneyApplication
		}
		if investmentRow.HasClass("advisory-fee") {
			anTransaction.Type = AdvisoryFee
			anTransaction.Investment = investmentRow.Find("td:nth-child(1) span").Text()
		}
		if len(investmentRow.Find("span.color-redemption").Nodes) > 0 {
			anTransaction.Type = Redemption
		}
		if len(investmentRow.Find("span.color-expired-asset").Nodes) > 0 {
			anTransaction.Type = ExpiredTitle
		}
		if len(investmentRow.Find("span.color-ir").Nodes) > 0 {
			anTransaction.Type = IRWithdrawal
		}

		applications = append(applications, anTransaction)
	}
	return applications, nil
}

func convertPtToEnNumber(investmentValue string) float64 {
	value, _ := strconv.ParseFloat(strings.TrimSpace(strings.Replace(strings.Replace(investmentValue, ".", "", -1), ",", ".", -1)), 64)
	return value
}
