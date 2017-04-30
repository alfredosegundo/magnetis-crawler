package magnetis

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"sort"
	"time"

	"net/url"

	"github.com/PuerkitoBio/goquery"
)

var host string = "https://magnetis.com.br"

type Equity struct {
	Time  time.Time
	Value string
}

type EquityCurve struct {
	Equities []Equity
}

func (e EquityCurve) Len() int      { return len(e.Equities) }
func (e EquityCurve) Swap(i, j int) { e.Equities[i], e.Equities[j] = e.Equities[j], e.Equities[i] }
func (e EquityCurve) Less(i, j int) bool {
	return e.Equities[i].Time.Before(e.Equities[j].Time)
}

type InvestmentPlan struct {
	Age               int
	Experience        string
	Goal              string
	GoalValue         float32 `json:"goal_value"`
	InitialInvestment float32 `json:"initial_investment"`
	LossTolerance     string  `json:"loss_tolerance"`
	MonthlyInvestment float32 `json:"monthly_investment"`
	PeriodInYears     int     `json:"period_in_years"`
	RiskLevel         int     `json:"risk_level"`
	RiskProfile       string  `json:"risk_profile"`
}

type Asset struct {
	Amount             string
	AssetId            int    `json:"asset_id"`
	AssetReturn        string `json:"asset_return"`
	CategoryKey        string `json:"category_key"`
	InstrumentTypeName string `json:"instrument_type_name"`
	Issuer             string
	Liquidity          int
	MaturityDate       string `json:"maturity_date"`
	Name               string
	Yield              string
}

var jar, _ = cookiejar.New(nil)
var defaultClient = &http.Client{
	Jar: jar,
}

func GetEquityCurve(userId string) (curve *EquityCurve, err error) {
	resp, err := defaultClient.Get(host + "/pricing/api/portfolio/" + userId + "/equity_curve")
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	icurve := make([][]interface{}, 0)

	err = json.Unmarshal(body, &icurve)
	if err != nil {
		return
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

func Signin(username string, password string) (err error) {
	var signin = host + "/users/sign_in"
	res, err := defaultClient.Get(signin)
	if err != nil {
		return err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
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
