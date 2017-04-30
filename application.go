package main

import (
	"fmt"
	"log"

	"MagnetisCrawler/magnetis"
	"os"
)

var userId string = os.Getenv("MAGNETIS_USER_ID")
var username string = os.Getenv("MAGNETIS_USER")
var password string = os.Getenv("MAGNETIS_PASS")

func main() {
	err := magnetis.Signin(username, password)
	if err != nil {
		log.Fatal(err)
	}

	curve, err := magnetis.GetEquityCurve(userId)
	if err != nil {
		log.Fatal(err)
	}

	equities := curve.Equities
	for i := range equities {
		fmt.Println(equities[i])
	}

	plan, err := magnetis.GetInvestmentPlan(userId)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(plan)

	assets, err := magnetis.Assets(userId)
	if err != nil {
		log.Fatal(err)
	}

	for i := range assets {
		fmt.Printf("%#v\n", assets[i])
	}
}
