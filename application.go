package main

import (
	"fmt"
	"log"

	"MagnetisCrawler/magnetis"
	"os"

	"github.com/urfave/cli"
)

var userId string = os.Getenv("MAGNETIS_USER_ID")
var username string = os.Getenv("MAGNETIS_USER")
var password string = os.Getenv("MAGNETIS_PASS")

func main() {
	var shouldSave bool
	var shouldPrint bool
	app := cli.NewApp()
	app.Name = "Magnetis Crawler"
	app.Usage = "Get my data form magnetis website"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "save, s",
			Usage:       "if we should save on drive",
			Destination: &shouldSave,
		},
		cli.BoolFlag{
			Name:        "print, p",
			Usage:       "Print on the console",
			Destination: &shouldPrint,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "curve",
			Aliases: []string{"c"},
			Usage:   "Get your equity curve from magnetis api",
			Action: func(c *cli.Context) error {
				err := magnetis.Signin(username, password)
				if err != nil {
					log.Fatal(err)
				}

				curve, err := magnetis.GetEquityCurve(userId)
				if err != nil {
					log.Fatal(err)
				}
				if shouldPrint {
					equities := curve.Equities
					for i := range equities {
						fmt.Println(equities[i])
					}
				}
				if shouldSave {
					log.Fatal("Not implemented yet")
				}
				return nil
			},
		},
		{
			Name:  "plan",
			Usage: "Get your investment plan from magnetis api",
			Action: func(c *cli.Context) error {
				err := magnetis.Signin(username, password)
				if err != nil {
					log.Fatal(err)
				}
				plan, err := magnetis.GetInvestmentPlan(userId)
				if err != nil {
					log.Fatal(err)
				}
				if shouldPrint {
					fmt.Printf("%#v\n", plan)
				}
				if shouldSave {
					log.Fatal("Not implemented yet")
				}
				return nil
			},
		},
		{
			Name:    "assets",
			Aliases: []string{"a"},
			Usage:   "Get your assets from magnetis api",
			Action: func(c *cli.Context) error {
				err := magnetis.Signin(username, password)
				if err != nil {
					log.Fatal(err)
				}
				assets, err := magnetis.Assets(userId)
				if err != nil {
					log.Fatal(err)
				}
				if shouldPrint {
					for i := range assets {
						fmt.Printf("%#v\n", assets[i])
					}
				}
				if shouldSave {
					log.Fatal("Not implemented yet")
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}
