package main

import (
	"flag"
	"fmt"
	"os"

	"autodeploy/cmd"
	"autodeploy/util/clean"

	"github.com/rs/zerolog/log"
)

func main() {
	flag.Parse()
	clean.Register(func() {

	})
	defer clean.Run()
	log.Info().Msg("autodeploy: is starting")
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Info().Msg("autodeploy: is end")

}
