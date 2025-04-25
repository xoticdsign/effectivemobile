package app

type App struct {
}

func New() App {
	return App{}
}

func (a App) Run() {}

func (a App) shutdown() {}
