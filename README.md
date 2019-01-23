Prometheus WebHook to SNMP-trap forwarder
=========================================

This is a quick (and dirty) way to get Prometheus to send SNMP traps, by mapping AlertManager "Annotations" and "Labels" to generic SNMP OIDs.

Integration with Prometheus
---------------------------
1. Prometheus gathers metrics
2. Prometheus appraises metrics against rules
3. If rules are triggered then alerts are raised through the AlertManager
4. The AlertManager triggers notifications to the webhook_snmptrapper
5. The webhook_snmptrapper forwards alerts as SNMP traps to the configured trap-address

SNMP integration
----------------
The provided MIB (`PROMETHEUS-TRAPPER-MIB.txt`) defines two notifications:
- ***prometheusTrapperFiringNotification***: Notification for an alert that has occurred
- ***prometheusTrapperRecoveryNotification***: Notification for an alert that has recovered

The MIB can be loaded into whatever SNMP Trap-server you're using. See [Dockerfile](trapdebug/net-snmp/Dockerfile) for a working demo using net-snmp on Alpine Linux.

### SNMP variables
Both of these traps contain the following variables:
- ***prometheusTrapperNotificationInstance***: The instance or hostname
- ***prometheusTrapperNotificationService***: A name for the service affected
- ***prometheusTrapperNotificationLocation***: The physical location where the alert was generated
- ***prometheusTrapperNotificationSeverity***: The severity of the alert
- ***prometheusTrapperNotificationDescription***: Text description of the alert
- ***prometheusTrapperNotificationTimestamp***: When the alert was first generated

AlertManager configuration
--------------------------
AlertManager needs to be configured to fire webhooks as notifications, with a pre-defined assortment of labels and annotations (these map to the SNMP MIB provided). Each alert should have the following parameters:

### Annotations:
- ***description***: A string describing the alert (_prometheusTrapperNotificationDescription_)
- ***info***: A string describing the alert (_prometheusTrapperNotificationInfo_)
- ***summary***: A string describing the alert (_prometheusTrapperNotificationSummary_)

### Labels:
- ***instance***: A string containing a unique host-identifier / hostname / instance-id / IP-address etc (_prometheusTrapperNotificationInstance_)
- ***severity***: A string describing the severity of the alert (_prometheusTrapperNotificationSeverity_)
- ***location***: A string describing the location of the instance(s) / system(s) generating the alert (_prometheusTrapperNotificationLocation_)
- ***service***: A string describing the service affected (_prometheusTrapperNotificationService_)

Command-line flags
------------------
- **--snmpcommunity**: The SNMP community string (_default_ = `public`)
- **--snmpversion**: SNMP protocol version to use (_default_ = `1`)
- **--snmpretries**: The number of times to retry sending traps (_default_ = `1`)
- **--snmptrapaddress**: The address to send traps to (_default_ = `127.0.0.1:162`)
- **--webhookaddress**: The address to listen for incoming webhooks on (_default_ = `0.0.0.0:9099`)


Testing
--------

````bash
#!/bin/bash
alerts='[
  {
    "labels": {
       "alertname": "instance_down",
       "instance": "example1"
     },
     "annotations": {
        "info": "The instance example1 is down",
        "summary": "instance example1 is down"
      }
  }
]'

URL="http://prometheus-sydxc-alertmanager.prometheus-sydxc.159.69.190.208.xip.io"

curl -XPOST -d"$alerts" $URL/api/v1/alerts

````

global:
inhibit_rules:
  - source_match:
      severity: 'warning'
    target_match:
      severity: 'normal'
    equal:
      - 'alertname'
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal:
      - 'alertname'
receivers:
  - name: 'snmp-forwarder'
    webhook_configs:
      - url: 'http://href.synology.me:9099'
        send_resolved: true
route:
  group_by:
    - alertname
    - host
  group_wait: '30s'
  group_interval: '5m'
  repeat_interval: '1h'
  receiver: 'snmp-forwarder'

