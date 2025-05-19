package main

import (
	"github.com/mlctrez/ehugo/service"
	"github.com/mlctrez/servicego"
)

func main() {
	servicego.Run(service.New())
}
