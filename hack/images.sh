#!/bin/bash
########################################################### {COPYRIGHT-TOP} ####
# Licensed Materials - Property of IBM
# 5900-AEO
#
# Copyright IBM Corp. 2020. All Rights Reserved.
#
# US Government Users Restricted Rights - Use, duplication, or
# disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
########################################################### {COPYRIGHT-END} ####

# Original from https://github.ibm.com/automation-base-pak/elastic-operator/blob/78210e859f10500baaa8e6658f2d23447e9ff51a/hack/images.sh

function=$1

images_file="$(dirname "$0")"/images.yaml
images=$(yq r -j "${images_file}" 'images.*')

parse_image() {
    image=${1}
    image_id=$(jq -r '.id' <<< "${image}")
    image_name=$(jq -r '.name' <<< "${image}")
    image_registry=$(jq -r '.registry' <<< "${image}")
    image_tag=$(jq -r '.tag' <<< "${image}")

    # flag used for fyre ocp dev, set image registry to internal and use tags in url
    if [ "${USE_OCP_REGISTRY}" == "true" ]; then
        image_registry=${INTERNAL_REGISTRY}
    fi
    image_url="${image_registry}/${image_name}:${image_tag}"
}


search_and_replace_images_for_kustomize() {
    read -r -d '' stdinput
    config=${stdinput}
    for image in ${images}; do
        parse_image "${image}"
        config=${config//\$\{${image_id}\}/${image_url}}
    done
    config=${config//\$\{NS\}/${NS}}
    echo "${config}"
}

$function
