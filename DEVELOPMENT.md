This document explains how to build and run the code example in this repo. See [README.md](./README.md) for context.

## Pre-requisites

### Required Binaries

- [ibmcloud >= v1.2.3](https://cloud.ibm.com/docs/cli?topic=cli-getting-started) - ibmcloud cli needed to authenticate and push images to ibm cloud registry.
- [go >= v1.15](https://golang.org/doc/install) - Go clients for talking to a Kubernetes cluster.
- [operator-sdk >= v1.0.1](https://sdk.operatorframework.io/docs/installation/install-operator-sdk/) - The Operator SDK is a framework that uses the controller-runtime library to make writing operators easier.
- [kustomize >= v3.8.0](https://github.com/kubernetes-sigs/kustomize/tags) - kustomize lets you customize raw, template-free YAML files for multiple purposes, leaving the original YAML untouched and usable as is.
- [opm >= v1.13.9](https://github.com/operator-framework/operator-registry/releases/tag/v1.13.9) - generates and updates registry databases as well as the index images that encapsulate them.
- [docker](https://docs.docker.com/engine/install/) - To build the container image.

**Note :** Validate all the required binaries are present and added to `path`. if not, set them to [path](https://opensource.com/article/17/6/set-path-linux). This is required to build and push the operator.

### Required local setup

**Environment Variables:**

- `API_KEY` - For pushing and pulling images from the IBM Cloud Registry (us.icr.io): [get your own API key here.](https://cloud.ibm.com/docs/account?topic=account-userapikey#create_user_key)

- `GIT_TOKEN` - For pulling code from other private registries on `github.ibm.com`: see [these instructions](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token)  to get your own.

- `GOPRIVATE="github.ibm.com"` - This is required for local Go builds to pull code from other private registries on `github.ibm.com`.

- `IMAGE_VERSION` - A unique tag for your builds, for example `mytest`. Your development versions of the images will be pushed with this tag. (Note: this must match what is in your `hack/images.yaml`)

If you have not already, you may need to run this command in order to get your local `go` tools fully working with the private code repositories:
```bash
git config --global url."https://git:<replace_with_git_token>@github.ibm.com/".insteadOf "https://github.ibm.com/"`
```

## Setup the Openshift cluster

If you have successfully installed IAF in the cluster already, uninstall it with the latest uninstall script in the [abp-deploy repository](https://github.ibm.com/automation-base-pak/abp-deploy). If you have not previously installed IAF, setup your cluster according to the instructions in the [abp-deploy repository](https://github.ibm.com/automation-base-pak/abp-deploy).

Run the latest setup script from the abp-deploy repository, for example: [install-iaf-setup.sh](https://github.ibm.com/automation-base-pak/abp-deploy/blob/main/install-latest/install-iaf-setup.sh). This will ensure that your cluster is set up properly, and create the required CatalogSources.

## Build and Deploy your Demo Cartridge

**a.** Adjust the `hack/images.yaml` to you match your `IMAGE_VERSION` environment variable. For example, if you have `IMAGE_VERSION='my-test-version'`, you would want this:
```yaml
images:
  - id: DEMO_CARTRIDGE
    registry: us.icr.io/abp-scratchpad
    name: iaf-demo-cartridge
    tag: my-test-version
```

**b.** Remove any existing resources for the demo cartridge (including the CatalogSource created by the abp-deploy setup script) with this command. If none was installed, it will have no effect.
```bash
make clean-demo
```

**c.** Build and Push the Operator, Bundle, and Catalog Images to the image registry with this command:
```bash
make build
```

**d.** Install the demo cartridge operator (by creating a CatalogSource and Subscription) and create an IAFDemo instance with this command:
```bash
make deploy-demo
```

A useful step to ensure everything works successfully is to do the following.

Ensure there is no data first in your ElasticSearch indices for iafdemo-raw and iaf-demo-anomaly.

```
curl -k -u elasticsearch-admin:<password> https://iaf-system-es-acme-abp.apps.<your cluster name.cp.fyre.ibm.com/<index name>/_count
```

If there is, you can delete all documents at a specific index with:

```
curl -k -u elasticsearch-admin:<password> -X DELETE https://iaf-system-es-acme-abp.apps.<your cluster name.cp.fyre.ibm.com/<index name>
```

After running the entire demo flow, use the same ElasticSearch query to count the number of documents in these indices, replacing the index name accordingly:

The password can be retrieved by base64 decoding the `password` value in the `iaf-system-elasticsearch-es-default-user` secret.

With a default configuration, assuming the data has been sent exactly once, you should have 1725 documents in the raw index and 82 documents in the anomaly index.

## Iterating

Ideally, any new changes should be able to be applied in the cluster with this command:
```bash
make clean-demo && make build && make deploy-demo
```

You can also simply do a `make build` and delete the EventProcessingTask, and then bounce the demo cartridge operator pod, if you are, for example, only testing out an Event Processing specific change (such as with the Flink job). Example:

```bash
oc delete eventprocessingtask iaf-eventprocessor-instance && oc delete pod -l app.kubernetes.io/name=iaf-demo-cartridge-operator
```

You may also be required to delete the Elastic user on repeated iterations (e.g. on the second, third, fourth, etc, as this is left-over from the previous deploy).

You can do this with

```
curl -k -X POST -u elasticsearch-admin:<the admin password> https://iaf-system-es-acme-abp.apps.<your cluster name>.cp.fyre.ibm.com/.security/_delete_by_query?pretty -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "username": "elasticsearch-admin"
    }
  }
}
'
```

whereby the password can be retrieved by base64 decoding the `password` value in the `iaf-system-elasticsearch-es-default-user` secret.

In some cases leftover resources may cause problems for repeated installations, in which case it is helpful to remove the entire installation with the [uninstall script from abp-deploy](https://github.ibm.com/automation-base-pak/abp-deploy/blob/main/install-latest/uninstall-iaf.sh), and then run the [setup script](https://github.ibm.com/automation-base-pak/abp-deploy/blob/main/install-latest/install-iaf-setup.sh), before running the `make` commands again.

## Architecture: The Producer

The producer pushes the provided [sample data](./pkg/producer/sample.csv) into the "iafdemo-raw" topic. The data is originally from https://ibm.box.com/s/tchm54j0azy86t2zxj61a86lpy9wfrbf

Each Kafka message will have a random UUID, CloudEvent headers, and will contain one row of the CSV as a labelled JSON. Here are a few examples (from kafkacat):

```
1613 ce_id=f0a4d41b-c5db-4519-8900-ce0a9cb8928f,ce_source=https://github.ibm.com/automation-base-pak/abp-demo-cartridge,ce_specversion=1.0,ce_type=bai.events.sample,content-type=application/json,ce_time=2018-12-07T16:39:00Z {"Req_Line_ID":"0010255080_10","Order_Line_ID":"4500265126_10","Goods_ID":"","Invoice_ID":"","Activity":"Order Line Created","DateTime":"2018-12-07T16:39:00Z","Resource":"DST25","Role":"Procurement","Requisition_Vendor":"","Order_Vendor":"VND05666","Invoice_Vendor":"","Pay_Vendor":"","Requisition_Type":"Direct","Order_Type":"Standard Order","Purchasing_Group":"ID2","Purchasing_Organization":"IT10","Material_Group":"D008-FOG","Material_Number":"108004509","Plant":"","Good_ReferenceNumber":"","Requisition_Header":"10255080","Order_Header":"4500265126","Invoice_Header":"","ClearDoc_Header":"","Good_Year":"","Invoice_Year":"","Order_Line_Amount":"49600","Invoice_Amount":"0","Paid_Amount":"0","Invoice_Document_Date":"","Invoice_Due_Date":"","Pay_Type":"","Pay_Delay":"0","UserType":"HUMAN","Invoice_Is_Overdue":""}
1624 ce_id=aafba3f4-5c27-42c3-8035-01414a306223,ce_source=https://github.ibm.com/automation-base-pak/abp-demo-cartridge,ce_specversion=1.0,ce_type=bai.events.sample,content-type=application/json,ce_time=2018-12-12T12:12:00Z {"Req_Line_ID":"","Order_Line_ID":"4500261624_30","Goods_ID":"5000941012_2_2018","Invoice_ID":"","Activity":"Goods Line Registered","DateTime":"2018-12-12T12:12:00Z","Resource":"DST55","Role":"Warehouse","Requisition_Vendor":"","Order_Vendor":"VND05137","Invoice_Vendor":"","Pay_Vendor":"","Requisition_Type":"","Order_Type":"Standard Order","Purchasing_Group":"II3","Purchasing_Organization":"IT11","Material_Group":"S140-0091","Material_Number":"S140009101","Plant":"IT01","Good_ReferenceNumber":"80266525","Requisition_Header":"","Order_Header":"4500261624","Invoice_Header":"","ClearDoc_Header":"","Good_Year":"2018","Invoice_Year":"","Order_Line_Amount":"395","Invoice_Amount":"0","Paid_Amount":"0","Invoice_Document_Date":"","Invoice_Due_Date":"","Pay_Type":"","Pay_Delay":"0","UserType":"HUMAN","Invoice_Is_Overdue":""}
1723 ce_id=2d33c532-22e3-4e0c-8ab8-02f73e9df4ca,ce_source=https://github.ibm.com/automation-base-pak/abp-demo-cartridge,ce_specversion=1.0,ce_type=bai.events.sample,content-type=application/json,ce_time=2019-03-18T12:45:00Z {"Req_Line_ID":"","Order_Line_ID":"","Goods_ID":"","Invoice_ID":"2519002133_2019_IT10","Activity":"Invoice Cleared","DateTime":"2019-03-18T12:45:00Z","Resource":"","Role":"Procurement","Requisition_Vendor":"","Order_Vendor":"","Invoice_Vendor":"VND06050","Pay_Vendor":"VND06050","Requisition_Type":"","Order_Type":"","Purchasing_Group":"","Purchasing_Organization":"","Material_Group":"","Material_Number":"","Plant":"","Good_ReferenceNumber":"","Requisition_Header":"","Order_Header":"","Invoice_Header":"2519002133","ClearDoc_Header":"6019020196","Good_Year":"","Invoice_Year":"2019","Order_Line_Amount":"0","Invoice_Amount":"1625","Paid_Amount":"1625","Invoice_Document_Date":"","Invoice_Due_Date":"2/28/19","Pay_Type":"Late","Pay_Delay":"18","UserType":"","Invoice_Is_Overdue":"yes"}
```

### Enabling the Knative connection to M2 (optional)

By default, the Knative connection is disabled. Unfortunately, the required Knative CRDs can not be installed automatically by ODLM, so Knative needs to be installed into the cluster manually. First, the Openshift Serverless operator needs to be installed. This can be accommplished through the OperatorHub, or by creating this Subscription manually:
```bash
oc apply -f - <<EOF
apiVersion: operators.coreos.com/v1
kind: Subscription
metadata:
  name: serverless-operator
  namespace: openshift-operators
spec:
  channel: "4.5"
  installPlanApproval: Automatic
  name: serverless-operator
  source: redhat-operators
  sourceNamespace: openshift-marketplace
  startingCSV: serverless-operator.v1.8.0
EOF
```

Once the Openshift Serverless operator is running, Knative Serving and Eventing must be instantiated (this will register their CRDs, eg `InMemoryChannel`):
```bash
oc apply -f - <<EOF
---
apiVersion: v1
kind: Namespace
metadata:
  name: knative-eventing
---
apiVersion: v1
kind: Namespace
metadata:
  name: knative-serving
---
apiVersion: operator.knative.dev/v1alpha1
kind: KnativeEventing
metadata:
  name: knative-eventing
  namespace: knative-eventing
---
apiVersion: operator.knative.dev/v1alpha1
kind: KnativeServing
metadata:
  name: knative-serving
  namespace: knative-serving
EOF
```

Finally, one additional Knative CRD is needed to connect to Kafka:
```bash
oc apply -f "https://github.com/knative/eventing-contrib/releases/download/v0.17.0/kafka-source.yaml"
```

Once all of that is installed in the cluster, the [manager](./config/manager/manager.yaml) can be updated to have `USE_KNATIVE=true`. That file can be edited here and then applied via `make deploy`, or the deployment can be edited inside the cluster. Once the manager pod has restarted, it should create the KafkaSource, InMemoryChannel, and Subscription that will forward *new* messages from the `iafdemo-anomaly` topic to the `demoserver` pod.

Note that only *new* messages will be sent to the demoserver pod. It may be necessary to restart the demoproducer pod or Flink job in order to populate new events. For example, a sample JSON was sent manually and can be seen in these logs:
```
$ oc logs demoserver-57dd796c97-z9p2z
Now listening...
POST / HTTP/1.1
Host: demoserver.abpdemo-knative-1.svc.cluster.local
Accept-Encoding: gzip
Ce-Id: partition:0/offset:0
Ce-Knativehistory: demoserver-kn-channel.abpdemo-knative-1.svc.cluster.local
Ce-Source: /apis/v1/namespaces/abpdemo-knative-1/kafkasources/demoserver#iafdemo-anomaly
Ce-Specversion: 1.0
Ce-Subject: partition:0#0
Ce-Time: 2020-10-02T18:17:31.362Z
Ce-Traceparent: 00-520b0355a34e136eb24ca634cccf5f1d-43a82ed534a8ccb5-00
Ce-Type: dev.knative.kafka.event
Content-Length: 13
Traceparent: 00-520b0355a34e136eb24ca634cccf5f1d-5af96cdc98f18f11-00
User-Agent: Go-http-client/1.1

{"foo":"bar"}
```
