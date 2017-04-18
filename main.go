package main

import (
	"github.com/reechou/robot-weishang/config"
	"github.com/reechou/robot-weishang/controller"
)

func main() {
	controller.NewLogic(config.NewConfig()).Run()
}
