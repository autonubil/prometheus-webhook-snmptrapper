package snmptrapper

import (
	"strconv"
	"strings"
	"time"

	types "github.com/autonubil/prometheus-webhook-snmptrapper/types"

	logrus "github.com/Sirupsen/logrus"
	snmpgo "github.com/autonubil/snmpgo"
	raven "github.com/getsentry/raven-go"
)

func getIpaddress(address string) (*snmpgo.Ipaddress, error) {
	addressParts := strings.Split(strings.Split(address, ":")[0], ".")
	res := make([]byte, len(addressParts))
	for idx, val := range addressParts {
		parsedVal, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		res[idx] = byte(parsedVal)
	}
	return snmpgo.NewIpaddress(res[0], res[1], res[2], res[3]), nil
}

func sendTrap(alert types.Alert) {

	// Prepare an SNMP handler:
	snmp, err := snmpgo.NewSNMP(snmpgo.SNMPArguments{
		Version:   myConfig.SNMPVersion,
		Address:   myConfig.SNMPTrapAddress,
		Retries:   myConfig.SNMPRetries,
		Community: myConfig.SNMPCommunity,
	})
	if err != nil {
		raven.CaptureError(err, map[string]string{})
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to create snmpgo.SNMP object")
		return
	} else {
		logrus.WithFields(logrus.Fields{"address": myConfig.SNMPTrapAddress, "retries": myConfig.SNMPRetries, "community": myConfig.SNMPCommunity}).Debug("Created snmpgo.SNMP object")
	}

	// Build VarBind list:
	var varBinds snmpgo.VarBinds
	var specificType int

	// The "enterprise OID" for the trap (rising/firing or falling/recovery):
	if alert.Status == "firing" {
		specificType = 1
		varBinds = append(varBinds, snmpgo.NewVarBind(snmpgo.OidSnmpTrap, trapOIDs.FiringTrap))
		varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.TimeStamp, snmpgo.NewOctetString([]byte(alert.StartsAt.Format(time.RFC3339)))))
	} else {
		specificType = 2
		varBinds = append(varBinds, snmpgo.NewVarBind(snmpgo.OidSnmpTrap, trapOIDs.RecoveryTrap))
		varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.TimeStamp, snmpgo.NewOctetString([]byte(alert.EndsAt.Format(time.RFC3339)))))
	}

	// Insert the AlertManager variables:
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Description, snmpgo.NewOctetString([]byte(alert.Annotations["description"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Info, snmpgo.NewOctetString([]byte(alert.Annotations["info"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Summary, snmpgo.NewOctetString([]byte(alert.Annotations["summary"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Instance, snmpgo.NewOctetString([]byte(alert.Labels["instance"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Severity, snmpgo.NewOctetString([]byte(alert.Labels["severity"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Location, snmpgo.NewOctetString([]byte(alert.Labels["location"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Service, snmpgo.NewOctetString([]byte(alert.Labels["service"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.JobName, snmpgo.NewOctetString([]byte(alert.Labels["job"]))))

	// Create an SNMP "connection":
	if err = snmp.Open(); err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to open SNMP connection")
		return
	}
	defer snmp.Close()

	// Send the trap:
	if myConfig.SNMPVersion == snmpgo.V1 {
		var enterprise *snmpgo.Oid
		var ipaddress *snmpgo.Ipaddress
		ipaddress, err = getIpaddress(alert.Address)
		if err == nil {
			enterprise, err = snmpgo.NewOid("1.3.6.1.4.1.39366.9093")
		}

		if err == nil {
			err = snmp.V1Trap(enterprise, 6, specificType, ipaddress, varBinds)
		}
	} else {
		err = snmp.V2Trap(varBinds)
	}
	if err != nil {
		raven.CaptureError(err, map[string]string{})

		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to send SNMP trap")
		return
	} else {
		if logrus.GetLevel() >= logrus.DebugLevel {
			logrus.WithFields(logrus.Fields{"status": alert.Status,
				"startsAt":     alert.StartsAt,
				"endsAt":       alert.EndsAt,
				"adress":       alert.Address,
				"annotations":  alert.Annotations,
				"generatorURL": alert.GeneratorURL,
				"Labels":       alert.Labels,
			}).Info("Send SNMP trap")

		} else {
			logrus.WithFields(logrus.Fields{"status": alert.Status}).Info("Send SNMP trap")
		}
	}
}
