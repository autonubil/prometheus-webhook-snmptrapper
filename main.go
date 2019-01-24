package main

import (
	"fmt"
	"os"

	raven "github.com/getsentry/raven-go"

	"github.com/autonubil/prometheus-webhook-snmptrapper/pkg/cmd"

	logrus "github.com/Sirupsen/logrus"
)

var Version string
var Commit string
var BuildDate string

var (
	log = logrus.WithFields(logrus.Fields{"logger": "main"})
)

func main() {
	// Set the log-level:
	logrus.SetLevel(logrus.InfoLevel)

	sentryDsn := os.Getenv("SENTRY_DSN")
	if len(sentryDsn) > 0 {
		raven.SetDSN(sentryDsn)
		raven.SetRelease(fmt.Sprintf("%s [%s@%s]", Version, Commit, BuildDate))
		// Make sure that the call to doStuff doesn't leak a panic
		raven.CapturePanic(run, nil)
	} else {
		run()
	}
}

func run() {
	// Create & execute new command
	cmd, err := cmd.NewSnmpTrapper()
	if err != nil {
		os.Exit(1)
	}

	log.Infof("Starting Prometheus SNMP Trapper [Version %s, Commit: %s, BuildDate: %s]", Version, Commit, BuildDate)
	raven.Capture(&raven.Packet{Level: raven.INFO, Message: "Starting Prometheus SNMP Trapper"}, map[string]string{"version": Version, "commit": Commit, "buildDate": BuildDate})

	err = cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)

}
