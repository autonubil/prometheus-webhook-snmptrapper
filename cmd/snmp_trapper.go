package cmd

/*
Copyright [2019] [autonubil System GmbH]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	goflag "flag"

	raven "github.com/getsentry/raven-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	config "github.com/autonubil/prometheus-webhook-snmptrapper/config"
	snmptrapper "github.com/autonubil/prometheus-webhook-snmptrapper/snmptrapper"
	types "github.com/autonubil/prometheus-webhook-snmptrapper/types"
	webhook "github.com/autonubil/prometheus-webhook-snmptrapper/webhook"

	logrus "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	cmdName   = "prometheus-webhook-snmptrapper"
	waitGroup = &sync.WaitGroup{}
	usage     = fmt.Sprintf("%s\n", cmdName)
)

// NewCmdOptions creates an options Cobra command to return usage
func NewCmdOptions() *cobra.Command {
	cmd := &cobra.Command{
		Use: "options",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	return cmd
}

// Create a new command for the grafanaConfig-operator. This cmd includes logging,
// cmd option parsing from flags, and the customization of the Tectonic assets.
func NewSnmpTrapper() (*cobra.Command, error) {
	conf := config.Config{}

	// Create a new command
	cmd := &cobra.Command{
		Use:   usage,
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			checkErr(Run(cmd, &conf), fatal)
		},
	}

	// Bind & parse flags defined by external projects.
	// e.g. This imports the golang/glog pkg flags into the cmd flagset
	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})

	cmd.Flags().BoolVarP(&conf.PrometheusEnabled, "prometheus", "p", conf.PrometheusEnabled, "Enable Prometheus metrics on port 9350.")

	// Process the command-line parameters:
	cmd.Flags().StringVarP(&conf.SNMPTrapAddress, "snmptrapaddress", "a", "127.0.0.1:162", "Address to send SNMP traps to")
	cmd.Flags().StringVarP(&conf.SNMPCommunity, "snmpcommunity", "c", "public", "SNMP community string")
	cmd.Flags().UintVarP(&conf.SNMPRetries, "snmpretries", "r", 1, "Number of times to retry sending SNMP traps")
	cmd.Flags().VarP(&config.SNMPVersionValue{SNMPVersion: &conf.SNMPVersion}, "snmpversion", "s", "SNMP protocol version")
	cmd.Flags().StringVarP(&conf.WebhookAddress, "webhookaddress", "w", "0.0.0.0:9099", "Address and port to listen for webhooks on")
	cmd.Flags().UintVarP(&conf.LogLevel, "loglevel", "v", 2, "Number of times to retry sending SNMP traps")

	// Set the log-level:
	if conf.LogLevel > 2 {
		logrus.SetLevel(logrus.DebugLevel)
	} else if conf.LogLevel == 2 {
		logrus.SetLevel(logrus.InfoLevel)
	} else if conf.LogLevel == 1 {
		logrus.SetLevel(logrus.WarnLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)

	}
	raven.SetTagsContext(map[string]string{
		"config.PrometheusEnabled": fmt.Sprintf("%v", conf.PrometheusEnabled),
		"config.LogLevel":          fmt.Sprintf("%v", conf.LogLevel),
		"config.SNMPCommunity":     fmt.Sprintf("%v", conf.SNMPCommunity),
		"config.SNMPRetries":       fmt.Sprintf("%v", conf.SNMPRetries),
		"config.SNMPTrapAddress":   fmt.Sprintf("%vS", conf.SNMPTrapAddress),
		"config.SNMPVersion":       fmt.Sprintf("%vS", conf.SNMPVersion),
		"config.WebhookAddress":    fmt.Sprintf("%vS", conf.WebhookAddress),
	})
	return cmd, nil
}

func serveMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9350", nil)
}

// Run the customization of the Tectonic assets
func Run(cmd *cobra.Command, conf *config.Config) error {

	if conf.PrometheusEnabled {
		go serveMetrics()
	}

	// Prepare a channel of events (to feed the digester):
	logrus.Info("Preparing the alerts channel")
	alertsChannel := make(chan types.Alert)

	// Prepare to have background GoRoutines running:
	waitGroup.Add(1)

	// Start webhook server:
	go webhook.Run(conf, alertsChannel, waitGroup)

	// Start the SNMP trapper:
	go snmptrapper.Run(conf, alertsChannel, waitGroup)

	// Make sure we wait for everything to complete before bailing out:
	waitGroup.Wait()

	raven.Capture(&raven.Packet{Level: raven.INFO, Message: "Stopped Prometheus SNMP Trapper"}, map[string]string{})

	return nil

}

func checkErr(err error, handleErr func(string)) {
	if err == nil {
		return
	}

	raven.CaptureError(err, map[string]string{"operation": "checkErr"})

	handleErr(err.Error())
}

// Fatal prints the message (if provided) and then exits. If V(2) or greater,
// glog.Fatal is invoked for extended information.
func fatal(msg string) {
	logrus.Fatal(msg)

	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(1)
}
