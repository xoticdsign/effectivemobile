package main

import "github.com/xoticdsign/effectivemobile/internal/app"

func main() {
	app, err := app.New()
	if err != nil {
		panic(err)
	}
	app.Run()
}
