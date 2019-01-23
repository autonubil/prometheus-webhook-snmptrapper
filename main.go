package main

import (
	flag "flag"
	sync "sync"

	config "github.com/autonubil/prometheus-webhook-snmptrapper/config"
	snmptrapper "github.com/autonubil/prometheus-webhook-snmptrapper/snmptrapper"
	types "github.com/autonubil/prometheus-webhook-snmptrapper/types"
	webhook "github.com/autonubil/prometheus-webhook-snmptrapper/webhook"

	logrus "github.com/Sirupsen/logrus"
)

var (
	conf      config.Config
	log       = logrus.WithFields(logrus.Fields{"logger": "main"})
	waitGroup = &sync.WaitGroup{}
)

func init() {
	// Process the command-line parameters:
	flag.StringVar(&conf.SNMPTrapAddress, "snmptrapaddress", "127.0.0.1:162", "Address to send SNMP traps to")
	flag.StringVar(&conf.SNMPCommunity, "snmpcommunity", "public", "SNMP community string")
	flag.UintVar(&conf.SNMPRetries, "snmpretries", 1, "Number of times to retry sending SNMP traps")
	flag.Var(&config.SNMPVersionValue{SNMPVersion: &conf.SNMPVersion}, "snmpversion", "SNMP protocol version")
	flag.StringVar(&conf.WebhookAddress, "webhookaddress", "0.0.0.0:9099", "Address and port to listen for webhooks on")
	flag.Parse()

	// Set the log-level:
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {

	// Make sure we wait for everything to complete before bailing out:
	defer waitGroup.Wait()

	// Prepare a channel of events (to feed the digester):
	log.Info("Preparing the alerts channel")
	alertsChannel := make(chan types.Alert)

	// Prepare to have background GoRoutines running:
	waitGroup.Add(1)

	// Start webhook server:
	go webhook.Run(conf, alertsChannel, waitGroup)

	// Start the SNMP trapper:
	go snmptrapper.Run(conf, alertsChannel, waitGroup)

}
