#!/usr/bin/env bash
########################################################### {COPYRIGHT-TOP} ####
# Licensed Materials - Property of IBM
# 5900-AEO
#
# Copyright IBM Corp. 2020. All Rights Reserved.
#
# US Government Users Restricted Rights - Use, duplication, or
# disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
########################################################### {COPYRIGHT-END} ####

files=(
  bundle.Dockerfile
  bundle/manifests/iaf-demo-cartridge-operator_v1_serviceaccount.yaml
  bundle/manifests/iaf-demo-cartridge.clusterserviceversion.yaml
  bundle/manifests/democartridge.ibm.com_iafdemoes.yaml
  bundle/metadata/annotations.yaml
  config/crd/bases/democartridge.ibm.com_iafdemoes.yaml
  config/manifests/bases/iaf-demo-cartridge.clusterserviceversion.yaml
  config/rbac/role.yaml
)

copyright_file="$(dirname ${0})/copyright.txt"

for file in ${files[@]}; do
  echo $file
  cat "$copyright_file" "$file" > temp && mv temp "$file"
done
