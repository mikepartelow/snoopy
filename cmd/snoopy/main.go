package main

import (
	"mp/snoopy/pkg/logging"

	"mp/snoopy/pkg/snoopy"
	"os"
)

func main() {
	logger, err := logging.Init()
	if err != nil {
		os.Stderr.WriteString("Error intializing logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	dogg, err := snoopy.New("config.yaml", logger)
	if err != nil {
		logger.Error("fatal", "error", err)
		os.Stderr.WriteString("Error intializing Snoopy: " + err.Error() + "\n")
		os.Exit(1)
	}

	dogg.Run()
}
