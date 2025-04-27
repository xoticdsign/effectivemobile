package main

import "github.com/xoticdsign/effectivemobile/internal/app"

// @title       EffectiveMobile
// @version     1.0.1
// @description Сервис предоставляет API для создания, обновления, удаления и поиска людей с автозаполнением данных на основе внешних сервисов.

// @contact.name  Епишов Роман
// @contact.url   https://github.com/xoticdsign
// @contact.email xoticdollarsign@outlook.com

// @license.name MIT
// @license.url  https://mit-license.org/

// @host     localhost:8080
// @BasePath /
// @accept   json
// @produce  json

// @schemes http
func main() {
	app, err := app.New()
	if err != nil {
		panic(err)
	}
	app.Run()
}
