######################################################### {COPYRIGHT-TOP} ###
# Licensed Materials - Property of IBM
# 5900-AEO
#
# Copyright IBM Corp. 2020, 2021
#
# US Government Users Restricted Rights - Use, duplication, or
# disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
######################################################### {COPYRIGHT-END} ###
FROM maven:3.6-openjdk-11 as builder
COPY src /app/src
COPY pom.xml /app
WORKDIR /app
RUN mvn install

FROM us.icr.io/abp-builds/iaf-flink
WORKDIR /opt/flink
COPY --from=builder --chown=1000:1000 /app/target/demo-flink-job-1.0-SNAPSHOT.jar demo/demo-flink-job-1.0-SNAPSHOT.jar
