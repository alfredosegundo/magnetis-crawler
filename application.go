package main

import (
	"fmt"
	"log"

	"github.com/alfredosegundo/magnetis-crawler/magnetis"
	"github.com/alfredosegundo/magnetis-crawler/spreadsheet"

	"os"

	"github.com/urfave/cli"
)

func main() {
	var userId string
	var username string
	var password string
	var shouldSave bool
	var shouldPrint bool

	app := cli.NewApp()
	app.Name = "Magnetis Crawler"
	app.Usage = "Get my data form magnetis website"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "userId, U",
			Usage:       "Your user id on magnetis api",
			Destination: &userId,
			EnvVar:      "MAGNETIS_USER_ID",
		},
		cli.StringFlag{
			Name:        "username, u",
			Usage:       "Your username on magnetis website",
			Destination: &username,
			EnvVar:      "MAGNETIS_USER",
		},
		cli.StringFlag{
			Name:        "password, p",
			Usage:       "Your password on magnetis api",
			Destination: &password,
			EnvVar:      "MAGNETIS_PASS",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "curve",
			Aliases: []string{"c"},
			Usage:   "Get your equity curve from magnetis api",
			Flags: []cli.Flag{
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
			},
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
					spreadsheet.Signin()
					spreadsheet.UpdateEquityCurve(curve)
				}
				return nil
			},
		},
		{
			Name:    "plan",
			Aliases: []string{"p"},
			Usage:   "Get your investment plan from magnetis api",
			Flags: []cli.Flag{
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
			},
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
			Flags: []cli.Flag{
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
			},
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
		{
			Name:    "applications",
			Aliases: []string{"ap"},
			Usage:   "Get your application history from magnetis website",
			Flags: []cli.Flag{
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
					fmt.Println(applications)
				}
				if shouldSave {
					spreadsheet.Signin()
					err = spreadsheet.UpdateApplications(applications)
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
