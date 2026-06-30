package main

import (
	"os"

	"github.com/mickmetalholic/dotfiles-bootstrap/internal/bootstrap"
)

func main() {
	os.Exit(bootstrap.Execute(os.Args[1:], bootstrap.Options{}))
}
