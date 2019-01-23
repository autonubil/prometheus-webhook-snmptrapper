package config

import (
	"errors"
	"strings"

	snmpgo "github.com/autonubil/snmpgo"
)

type SNMPVersionValue struct {
	SNMPVersion *snmpgo.SNMPVersion
}

func (v SNMPVersionValue) Type() string {
	return "SNMPVersionValue"
}

func (v SNMPVersionValue) String() string {
	if v.SNMPVersion != nil {
		switch *v.SNMPVersion {
		case snmpgo.V1:
			return "V1"
		case snmpgo.V2c:
			return "V2c"
		case snmpgo.V3:
			return "V3"

		}
	}
	return "V1"
}

func (v SNMPVersionValue) Set(s string) error {

	if s == "" {
		*v.SNMPVersion = 0
	} else {
		switch strings.ToLower(s) {
		case "":
			*v.SNMPVersion = 0
		case "1":
			*v.SNMPVersion = 0
		case "v1":
			*v.SNMPVersion = 0
		case "2":
			*v.SNMPVersion = 1
		case "v2":
			*v.SNMPVersion = 1
		case "v2c":
			*v.SNMPVersion = 1
		case "3":
			*v.SNMPVersion = 2
		case "v3":
			*v.SNMPVersion = 2
		default:
			return errors.New("Illegal version (allowed options: V1,V2c,V3)")
		}
	}

	return nil
}

type Config struct {
	SNMPTrapAddress   string
	SNMPCommunity     string
	SNMPRetries       uint
	WebhookAddress    string
	SNMPVersion       snmpgo.SNMPVersion
	PrometheusEnabled bool
	LogLevel          uint
}
