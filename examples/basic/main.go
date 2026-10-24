package main

import "github.com/beratkara/goralens/examples/basic/bootstrap"

func main() {
	app := bootstrap.Boot()
	app.Start()
}
