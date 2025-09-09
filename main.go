package main

import (
	"fmt"

	"github.com/alkeshnikalje/gator-go/internal/config"
)


func main () {
	cfg := config.Read()
	cfg.SetUser("Alkesh")
	cfgAgain := config.Read()
	fmt.Println(cfgAgain)
}
