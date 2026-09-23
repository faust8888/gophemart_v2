// Gophermart — HTTP-сервис накопительной системы лояльности «Гофермарт».
package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/faust8888/gophemart_v2/internal/app"
	"github.com/faust8888/gophemart_v2/internal/config"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatal(err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
