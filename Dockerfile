######################################################### {COPYRIGHT-TOP} ###
# Licensed Materials - Property of IBM
# 5900-AEO
#
# Copyright IBM Corp. 2020, 2021
#
# US Government Users Restricted Rights - Use, duplication, or
# disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
######################################################### {COPYRIGHT-END} ###
# Build the manager binary
FROM golang:1.13 as builder

ENV GOPRIVATE="github.ibm.com"
ARG GIT_TOKEN
RUN git config --global url."https://git:${GIT_TOKEN}@github.ibm.com/".insteadOf "https://github.ibm.com/"

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

FROM us.icr.io/abp-ubi8/iaf-minimal

WORKDIR /
COPY --from=builder /workspace/manager .
COPY pkg/producer/sample.csv .
#Copy AI Models
COPY models/ models/
RUN chmod g+w /var/log
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
