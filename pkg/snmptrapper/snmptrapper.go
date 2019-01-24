package snmptrapper

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	config "github.com/autonubil/prometheus-webhook-snmptrapper/pkg/config"
	types "github.com/autonubil/prometheus-webhook-snmptrapper/pkg/types"

	logrus "github.com/Sirupsen/logrus"
	snmpgo "github.com/autonubil/snmpgo"
	raven "github.com/getsentry/raven-go"
)

var (
	myConfig *config.Config
	trapOIDs types.TrapOIDs
)

func init() {
	// Set the log-level:
	logrus.SetLevel(logrus.DebugLevel)

	// Configure which OIDs to use for the SNMP Traps:
	trapOIDs.FiringTrap, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.0.1")
	trapOIDs.RecoveryTrap, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.0.2")
	trapOIDs.Instance, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.1.1")
	trapOIDs.Service, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.1.2")
	trapOIDs.Location, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.1.3")
	trapOIDs.Severity, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.1.4")
	trapOIDs.Description, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.1.5")
	trapOIDs.JobName, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.1.6")
	trapOIDs.TimeStamp, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.1.7")
	trapOIDs.Info, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.1.8")
	trapOIDs.Summary, _ = snmpgo.NewOid("1.3.6.1.4.1.39366.9093.1.9")
}

func Run(myConfigFromMain *config.Config, alertsChannel chan types.Alert, waitGroup *sync.WaitGroup) {

	logrus.WithFields(logrus.Fields{"address": myConfigFromMain.SNMPTrapAddress}).Info("Starting the SNMP trapper")

	// Populate the config:
	myConfig = myConfigFromMain

	// Set up a channel to handle shutdown:
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Kill, os.Interrupt)

	// Handle incoming alerts:
	go func() {
		for {
			select {

			case alert := <-alertsChannel:
				raven.Capture(&raven.Packet{Level: raven.DEBUG, Message: "New Alert"}, map[string]string{
					"status":       fmt.Sprintf("%v", alert.Status),
					"startsAt":     fmt.Sprintf("%v", alert.StartsAt),
					"endsAt":       fmt.Sprintf("%v", alert.EndsAt),
					"adress":       fmt.Sprintf("%v", alert.Address),
					"annotations":  fmt.Sprintf("%v", alert.Annotations),
					"generatorURL": fmt.Sprintf("%v", alert.GeneratorURL),
					"Labels":       fmt.Sprintf("%v", alert.Labels),
				})
				// Send a trap based on this alert:
				logrus.WithFields(logrus.Fields{"status": alert.Status}).Debug("Received an alert")
				sendTrap(alert)
			}
		}
	}()

	// Wait for shutdown:
	for {
		select {
		case <-signals:
			logrus.Warn("Shutting down the SNMP trapper")

			// Tell main() that we're done:
			waitGroup.Done()
			return
		}
	}

}
