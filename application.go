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

	"github.com/PuerkitoBio/goquery"
)

var host string = "https://magnetis.com.br"
var signin string = host + "/users/sign_in"
var equityCurve string = host + "/pricing/api/portfolio/20146/equity_curve"
var username string = os.Getenv("MAGNETIS_USER")
var password string = os.Getenv("MAGNETIS_PASS")

type Equity struct {
	Time  float64
	Value string
}

type EquityCurve struct {
	Curve []Equity
}

func (e EquityCurve) Len() int           { return len(e.Curve) }
func (e EquityCurve) Swap(i, j int)      { e.Curve[i], e.Curve[j] = e.Curve[j], e.Curve[i] }
func (e EquityCurve) Less(i, j int) bool { return e.Curve[i].Time < e.Curve[j].Time }

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
		fmt.Println(string(body))
		log.Fatal(err)
	}

	equityCurve := new(EquityCurve)
	for i := range curve {
		value := curve[i][1]
		equity := Equity{Time: curve[i][0].(float64), Value: value.(string)}
		equityCurve.Curve = append(equityCurve.Curve, equity)
	}

	sort.Sort(equityCurve)
	fmt.Println(equityCurve)
}
