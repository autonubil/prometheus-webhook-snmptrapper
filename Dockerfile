ARG REPO=github.com/autonubil/prometheus-webhook-snmptrapper
ARG VERSION=v0.0.0-debug
ARG BUILD_DATE=latest
ARG COMMIT=latest
FROM golang:1.10.2 AS build-env
ARG REPO
LABEL maintainer="carsten.zeumer@autonubil.de"

WORKDIR /usr/local/go/src

COPY pkg /usr/local/go/src/$REPO/pkg
COPY vendor /usr/local/go/src/$REPO/vendor
copy *.go /usr/local/go/src/$REPO/

WORKDIR /usr/local/go/src/$REPO

RUN pwd && ls -alh

RUN  CGO_ENABLED=0 GOOS=linux go build -ldflags "-X cmd/prometheus-webhook-snmptrapper.Version=$CI_COMMIT_REF_NAME -X cmd/prometheus-webhook-snmptrapper.BuildDate=$(date --iso-8601=seconds) -X cmd/prometheus-webhook-snmptrapper.Commit=$CI_COMMIT_SHA -s" -a -installsuffix cgo  -v -o /bin/prometheus-webhook-snmptrapper ./main.go


# final stage
FROM alpine:3.8
LABEL maintainer="carsten.zeumer@autonubil.de"

WORKDIR /

COPY --from=build-env /bin/prometheus-webhook-snmptrapper /prometheus-webhook-snmptrapper
CMD chmod +x /prometheus-webhook-snmptrapper

ENV SNMP_COMMUNITY="public"
ENV SNMP_RETRIES=1
ENV SNMP_VERSION="v1"
ENV SNMP_TRAP_ADDRESS="localhost:162"
ENV WEBHOOK_ADDRESS="0.0.0.0:9099"
ENV LOGLEVEL=2
ENV PROMETHEUS="true"

EXPOSE 9099
EXPOSE 9350

CMD ["/bin/sh", "-c", "/prometheus-webhook-snmptrapper --loglevel=$LOGLEVEL --prometheus=$PROMETHEUS --snmpversion=$SNMP_VERSION --snmpcommunity=$SNMP_COMMUNITY --snmpretries=$SNMP_RETRIES --snmptrapaddress=$SNMP_TRAP_ADDRESS --webhookaddress=$WEBHOOK_ADDRESS"]


