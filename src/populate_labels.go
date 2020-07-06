package main

import (
	"flag"
	clients "github.com/bbernhard/imagemonkey-core/clients"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	log "github.com/sirupsen/logrus"
)

func main() {
	dryRun := flag.Bool("dryrun", true, "dry run")
	debug := flag.Bool("debug", false, "debug")

	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *dryRun {
		log.Info("Populating labels (dry run)...")
	} else {
		log.Info("Populating labels...")
	}

	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")

	labelsPopulator := clients.NewLabelsPopulatorClient(imageMonkeyDbConnectionString,
		"../wordlists/en/labels.jsonnet",
		"../wordlists/en/label-refinements.json",
		"../wordlists/en/metalabels.jsonnet",
		"../wordlists/en/label-joints.json")
	err := labelsPopulator.Load()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = labelsPopulator.Populate(*dryRun)
	if err != nil {
		log.Fatal(err.Error())
	}
}
