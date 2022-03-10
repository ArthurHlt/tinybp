package main

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile = kingpin.Flag("config", "Configuration File").Default("config.yml").Short('c').String()
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	kingpin.Version(fmt.Sprintf("tinybp %s, commit %s, built at %s by %s", version, commit, date, builtBy))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	cnf, err := LoadConfig(*configFile)
	if err != nil {
		logrus.Fatal(err.Error())
		return
	}
	http.Handle("/traefik", NewTraefikHandler(cnf.Domain, cnf.Bookmarks))
	logrus.Infof("listen on: http://%s", cnf.Listen)
	logrus.Fatal(http.ListenAndServe(cnf.Listen, nil))
}
