package main

import "github.com/turnerlabs/harbor-compose/cmd"

var version string
var buildDate string

func main() {
	cmd.Execute(version, buildDate)
}
