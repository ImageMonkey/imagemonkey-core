package main

import(
	log "github.com/sirupsen/logrus"
	"flag"
	commons "github.com/bbernhard/imagemonkey-core/commons"
)

func main(){
	dryRun := flag.Bool("dryrun", true, "dry run")
	debug := flag.Bool("debug", false, "debug")

	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *dryRun {
		log.Info("Populating labels (dry run)...")
	} else{
		log.Info("Populating labels...")
	}

	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")
	
	labelsPopulator := commons.NewLabelsPopulator(imageMonkeyDbConnectionString, "../wordlists/en/labels.jsonnet", "../wordlists/en/label-refinements.json", "../wordlists/en/metalabels.jsonnet")
	err := labelsPopulator.Load()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = labelsPopulator.Populate(*dryRun)
	if err != nil {
		log.Fatal(err.Error())
	}
}
