package main

import (
	"fmt"
	"log"

	"net/http"
	"net/http/cookiejar"
	"net/url"

	"io/ioutil"

	"os"

	"encoding/json"

	"sort"

	"time"

	"github.com/PuerkitoBio/goquery"
)

var userId string = os.Getenv("MAGNETIS_USER_ID")
var username string = os.Getenv("MAGNETIS_USER")
var password string = os.Getenv("MAGNETIS_PASS")

var host string = "https://magnetis.com.br"
var signin string = host + "/users/sign_in"
var equityCurve string = host + "/pricing/api/portfolio/" + userId + "/equity_curve"
var plan string = host + "/api/investment_plan/" + userId

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
	Age               int `json:"age"`
	Experience        string `json:"experience"`
	Goal              string `json:"experience"`
	GoalValue         float32 `json:"goal_value"`
	InitialInvestment float32 `json:"initial_investment"`
	LossTolerance     string `json:"loss_tolerance"`
	MonthlyInvestment float32 `json:"monthly_investment"`
	PeriodInYears     int `json:"period_in_years"`
	RiskLevel         int `json:"risk_level"`
	RiskProfile       string `json:"risk_profile"`
}

func main() {
	cookieJar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: cookieJar,
	}
	res, err := client.Get(signin)
	if err != nil {
		log.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Fatal(err)
	}
	selection := doc.Find("input[name='authenticity_token']")
	token, _ := selection.First().Attr("value")
	client.PostForm(signin, url.Values{
		"authenticity_token": {token},
		"utf8":               {"✓"},
		"user[email]":        {username},
		"user[password]":     {password},
	})
	resp, err := client.Get(equityCurve)
	body, err := ioutil.ReadAll(resp.Body)

	curve := make([][]interface{}, 0)

	err = json.Unmarshal(body, &curve)
	if err != nil {
		log.Fatal(err)
	}

	equityCurve := new(EquityCurve)
	for i := range curve {
		equityTime := int64(curve[i][0].(float64)) / 1000
		equity := Equity{Time: time.Unix(equityTime, 0).UTC(), Value: curve[i][1].(string)}
		equityCurve.Equities = append(equityCurve.Equities, equity)
	}
	sort.Sort(equityCurve)

	equities := equityCurve.Equities
	for i := range equities {
		fmt.Println(equities[i])
	}

	resp, err = client.Get(plan)

	investment := new(InvestmentPlan)
	body, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &investment)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(investment)
}
