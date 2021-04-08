#!/bin/bash
########################################################### {COPYRIGHT-TOP} ####
# Licensed Materials - Property of IBM
# 5900-AEO
#
# Copyright IBM Corp. 2020, 2021. All Rights Reserved.
#
# US Government Users Restricted Rights - Use, duplication, or
# disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
########################################################### {COPYRIGHT-END} ####

echo -e "     ====> Configuring Image Registry Started...."
# echo $1

ibmcloud config --check-version=false
ibmcloud login --apikey $1 -a https://cloud.ibm.com -r us-south -g abp-ci-cd

ibmcloud cr region-set us-south
ibmcloud cr login
