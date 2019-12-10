package main

import (
	"fmt"
	"log"

	"os"

	"github.com/alfredosegundo/magnetis-crawler/magnetis"
	"github.com/alfredosegundo/magnetis-crawler/spreadsheet"

	"github.com/urfave/cli/v2"
)

func main() {
	var userID string
	var username string
	var password string
	var spreadsheetID string
	var shouldSave bool
	var shouldPrint bool
	var shouldPrintExcel bool

	app := cli.NewApp()
	app.Name = "Magnetis Crawler"
	app.Usage = "Get my data form magnetis website"
	app.Version = "1.0.2"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "userID",
			Aliases:     []string{"U"},
			Usage:       "Your user id on magnetis api",
			Destination: &userID,
			EnvVars:     []string{"MAGNETIS_USER_ID"},
		},
		&cli.StringFlag{
			Name:        "username",
			Aliases:     []string{"u"},
			Usage:       "Your username on magnetis website",
			Destination: &username,
			EnvVars:     []string{"MAGNETIS_USER"},
		},
		&cli.StringFlag{
			Name:        "password",
			Aliases:     []string{"p"},
			Usage:       "Your password on magnetis api",
			Destination: &password,
			EnvVars:     []string{"MAGNETIS_PASS"},
		},
		&cli.StringFlag{
			Name:        "spreadsheet, sheet",
			Aliases:     []string{"sheet"},
			Usage:       "Your spreadsheet id on google drive",
			Destination: &spreadsheetID,
			EnvVars:     []string{"GOOGLE_SPREADSHEET_ID"},
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:    "curve",
			Aliases: []string{"c"},
			Usage:   "Get your equity curve from magnetis api",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "save",
					Aliases:     []string{"s"},
					Usage:       "if we should save on drive",
					Destination: &shouldSave,
				},
				&cli.BoolFlag{
					Name:        "print",
					Aliases:     []string{"p"},
					Usage:       "Print on the console",
					Destination: &shouldPrint,
				},
				&cli.BoolFlag{
					Name:        "excel",
					Aliases:     []string{"e"},
					Usage:       "Print on the console as tab separated execel formated values",
					Destination: &shouldPrintExcel,
				},
			},
			Action: func(c *cli.Context) error {
				err := magnetis.Signin(username, password)
				if err != nil {
					log.Fatal(err)
				}

				curve, err := magnetis.GetEquityCurve(userID)
				if err != nil {
					log.Fatalf("Error retrieving equity curve: %v", err)
				}
				if shouldPrint {
					equities := curve.Equities
					for i := range equities {
						fmt.Println(equities[i])
					}
				}
				if shouldPrintExcel {
					equities := curve.Equities
					for i := range equities {
						fmt.Println(equities[i].Excel())
					}
				}
				if shouldSave {
					spreadsheet.SpreadsheetsSignin()
					err = spreadsheet.UpdateEquityCurve(curve.Equities, spreadsheetID)
					if err != nil {
						log.Fatal(err)
					}
				}
				return nil
			},
		},
		{
			Name:    "plan",
			Aliases: []string{"p"},
			Usage:   "Get your investment plan from magnetis api",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "save",
					Aliases:     []string{"s"},
					Usage:       "if we should save on drive",
					Destination: &shouldSave,
				},
				&cli.BoolFlag{
					Name:        "print",
					Aliases:     []string{"p"},
					Usage:       "Print on the console",
					Destination: &shouldPrint,
				},
			},
			Action: func(c *cli.Context) error {
				err := magnetis.Signin(username, password)
				if err != nil {
					log.Fatal(err)
				}
				plan, err := magnetis.GetInvestmentPlan(userID)
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
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "save",
					Aliases:     []string{"s"},
					Usage:       "if we should save on drive",
					Destination: &shouldSave,
				},
				&cli.BoolFlag{
					Name:        "print",
					Aliases:     []string{"p"},
					Usage:       "Print on the console",
					Destination: &shouldPrint,
				},
			},
			Action: func(c *cli.Context) error {
				err := magnetis.Signin(username, password)
				if err != nil {
					log.Fatal(err)
				}
				assets, err := magnetis.Assets(userID)
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
		{
			Name:    "applications",
			Aliases: []string{"ap"},
			Usage:   "Get your application history from magnetis website",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "save",
					Aliases:     []string{"s"},
					Usage:       "if we should save on drive",
					Destination: &shouldSave,
				},
				&cli.BoolFlag{
					Name:        "print",
					Aliases:     []string{"p"},
					Usage:       "Print on the console",
					Destination: &shouldPrint,
				},
				&cli.BoolFlag{
					Name:        "excel",
					Aliases:     []string{"e"},
					Usage:       "Print on the console as tab separated execel formated values",
					Destination: &shouldPrintExcel,
				},
			},
			Action: func(c *cli.Context) error {
				err := magnetis.Signin(username, password)
				if err != nil {
					log.Fatal(err)
				}
				applications, err := magnetis.Applications()
				if err != nil {
					log.Fatal(err)
				}
				if shouldPrint {
					for i := range applications {
						fmt.Println(applications[i])
					}
				}
				if shouldPrintExcel {
					for i := range applications {
						fmt.Println(applications[i].Excel())
					}
				}
				if shouldSave {
					spreadsheet.SpreadsheetsSignin()
					err = spreadsheet.UpdateApplications(applications, spreadsheetID)
					if err != nil {
						log.Fatal(err)
					}
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}
