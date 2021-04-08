######################################################### {COPYRIGHT-TOP} ###
# Licensed Materials - Property of IBM
# 5900-AEO
#
# Copyright IBM Corp. 2020, 2021. All Rights Reserved.
#
# US Government Users Restricted Rights - Use, duplication, or
# disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
######################################################### {COPYRIGHT-END} ###
FROM scratch

LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=iaf-demo-cartridge
LABEL operators.operatorframework.io.bundle.channels.v1=alpha
LABEL operators.operatorframework.io.metrics.builder=operator-sdk-v1.2.0
LABEL operators.operatorframework.io.metrics.mediatype.v1=metrics+v1
LABEL operators.operatorframework.io.metrics.project_layout=go.kubebuilder.io/v2
LABEL operators.operatorframework.io.test.config.v1=tests/scorecard/
LABEL operators.operatorframework.io.test.mediatype.v1=scorecard+v1
COPY bundle/manifests /manifests/
COPY bundle/metadata /metadata/
