FROM alpine:latest
MAINTAINER Prawn
USER root

RUN apk update
RUN apk add curl

ENV SNMP_COMMUNITY="public"
ENV SNMP_RETRIES=1
ENV SNMP_TRAP_ADDRESS="localhost:162"
ENV WEBHOOK_ADDRESS="0.0.0.0:9099"

EXPOSE 9099

COPY prometheus-webhook-snmptrapper /usr/local/bin/prometheus-webhook-snmptrapper
COPY sample-alert.json /

CMD exec /usr/local/bin/prometheus-webhook-snmptrapper -snmpcommunity=$SNMP_COMMUNITY -snmpretries=$SNMP_RETRIES -snmptrapaddress=$SNMP_TRAP_ADDRESS -webhookaddress=$WEBHOOK_ADDRESS

# docker build -t "prawn/prometheus-webhook-snmptrapper" .
