  - [iaf-demo-cartridge](#iaf-demo-cartridge)
  - [Pre-requisites](#pre-requisites)
  - [To start with](#to-start-with)
  - [Build Operator and Bundle Images](#build-operator-and-bundle-images)
  - [Deploy the iaf-demo-cartridge operator](#deploy-the-iaf-demo-cartridge-operator)
  - [The producer](#the-producer)
  - [Test-running the producer](#test-running-the-producer)
  - [Troubleshooting](#troubleshooting)
  - [For Developers](#for-developers)

# iaf-demo-cartridge

This repo contains an example Operator which depends on and uses IBM Automation Foundation services.
It is implemented using [operator-sdk](https://sdk.operatorframework.io/) as a Go operator, like many existing
IBM Cloud Paks.

This repo contains multiple components:
  - the Go operator (which registers itself as `IAF Demo Cartridge`) whose code is in the root folder of this repo
  - a separate Flink processor container, written in Java.  That code is described in [demo-flink-processor README](demo-flink-processor/README.md)
  - an AI model built with TensorFlow and Deployed using AI Foundation model serving.  This is described here [AI components in Demo Cartridge](doc/IAFAIREADME.md)

It demonstrates how to use specific IBM Automation Foundation capabilities to show an event
processing scenario simulating parts of the Procure to Pay scenario:
  - it has an OLM dependency on the IBM Automation Foundation custom resources, so that IAF is installed automatically if needed. This is set in [the ClusterServiceVersion file](bundle/manifests/iaf-demo-cartridge.clusterserviceversion.yaml).
  - it has its own Custom Resource type which can be created to start the "demo scenario" process
  - this sequence of processing is in [iafdemo_controller.go](controllers/iafdemo_controller.go)
  - when its Custom Resource is created, it registers as a `Cartridge` and waits for the relevant IAF resources to be created
  - once it has registered, it also creates two `EventStream` resources
  - it sets up a Flink job to run against those events, spot patterns, and re-emit them on another topic
  - it runs a container which sends formatted events to one of the Kafka topics it has set up
  - it runs another container which consumes those events to show they are being processed

## Running and using the Demo cartridge

There are automated builds of this repo run by the IBM Automation Foundation team, which are published as their own `CatalogSource` image into Staging Entitled Registry.  To use those pre-built images you need to follow the [IBM Automation Foundation playbook](https://pages.github.ibm.com/automation-base-pak/abp-playbook/planning-install/installing#obtaining-the-images) for registry setup steps.  The steps here also assume that you have already applied the IBM Automation Foundation catalogsource to your cluster - if not, follow the playbook again (same link).

Once you have done that one-time setup, install the Demo catalog source into a cluster using the pre-built images by creating the following CatalogSource:
```bash
cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1
kind: CatalogSource
metadata:
  name: demo-cartridge
  namespace: openshift-marketplace
spec:
  displayName: IAF Demo Cartridge
  publisher: IBM
  sourceType: grpc
  image: cp.stg.icr.io/cp/abp-demo-cartridge-catalog:0.0.11
  updateStrategy:
    registryPoll:
      interval: 45m
EOF
```

Verify that the CatalogSource was added and its pod is successfully running:
```bash
$ oc get pods -n openshift-marketplace | grep demo-cartridge
iaf-demo-cartridge-7lczk 1/1 Running 0 32s
```

Once that is available, you can run the Operator by creating a Subscription like this (note this depends on an OperatorGroup already existing - or you can create a Subscription by installing in the OpenShift console instead):

If there is not already an OperatorGroup in the project, create one like this:
```bash
cat <<EOF | oc apply -f -

apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: iaf-group
  namespace: ${IAF_PROJECT}
spec:
  targetNamespaces:
  - ${IAF_PROJECT}
EOF
```

The subscription is created with this custom resource:
```bash
cat <<EOF | oc apply -f -

apiVersion: operators.coreos.com/v1
kind: Subscription
metadata:
  name: iaf-demo-cartridge
  namespace: ${IAF_PROJECT}
spec:
  channel: alpha
  installPlanApproval: Automatic
  name: iaf-demo-cartridge
  source: demo-cartridge
  sourceNamespace: openshift-marketplace
EOF
```
Note that this installs IBM Automation Foundation and Common Services operators into the namespace `${IAF_PROJECT}` if they weren't already present.

Finally, create an instance of the `IAFDemo` custom resource, managed by this operator, to actually run the demo scenario:
```bash
cat <<EOF | oc apply -f -  

apiVersion: democartridge.ibm.com/v1
kind: IAFDemo
metadata:
  name: iafdemo-sample
  namespace: ${IAF_PROJECT}
spec:
  messagesPerGroup: "10"
  secondsToPause: "1"
  zen: true
  license:
    accept: true
EOF
```

Once this process completes, you should observe a pod called `demoproducer` in your namespace, which should show in its logs that it is sending Kafka messages. If the scenario is working then those events will be ingested into the Elasticsearch instance hosted by Automation Foundation. For example:

```bash
$ oc get pods -n $IAF_PROJECT | grep demoproducer
demoproducer-65bfdf588d-zdxcx       1/1     Running   0    110s
```

After some time, the eventprocessingtask should run: it will copy the 'raw' events into an elasticsearch index directly, and it will identify some 'anomalous' events which it will send to an 'anomaly' index. To check if those events are present in elasticsearch, you can use an elasticserach API to inspect them. First, you need to obtain the route to the elasticsearch instance, by using the following:

```bash
$ oc get route iaf-system-es -o=jsonpath='{.spec.host}'
iaf-system-es-acme-iaf.henry-cluster-lon02-b3c-dff...00.eu-gb.containers.appdomain.cloud
```

Before you can call an elasticsearch API using this route, you will need the username and password of the elasticsearch instance, set up by the IBM Automation Foundation. These are stored in the `iaf-system-elasticsearch-es-default-user` secret. You can obtain these with the following command:

```bash
$ oc extract secret/iaf-system-elasticsearch-es-default-user --to=-
# password
asC9.....socT92N
# username
elasticsearch-admin
```

To read the events, you can use the search API (utilizing the url, username and passowrd obtained above), which will return some summary information plus, by default, the first 10 documents in an index:

```bash
curl -k -u <username>:<password> <url>/iafdemo-raw-new/_search
```

You should see a total of 1725 hits for the search, i.e. something like this:

```bash
$ curl -k -u elasticsearch-admin:asC9.....socT92N https://iaf-system-es-acme-iaf.henry-cluster-lon02-b3c-dff...00.eu-gb.containers.appdomain.cloud/iafdemo-raw-new/_search
{
  "took": 23,
  "timed_out": false,
  "_shards": {
    "total": 1,
    "successful": 1,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": {
      "value": 1725,
      "relation": "eq"
    },
    "max_score": 1,
    "hits": [
      {
        "_index": "iafdemo-raw-new",
        "_type": "_doc",
        "_id": "uvFKAXgByKTPnu0afaWY",
        "_score": 1,
        "_source": {
          "Good_ReferenceNumber": "",
          "..."
        }
      },
      {
        "_index": "iafdemo-raw-new",
        "_type": "_doc",
        "_id": "u_FKAXgByKTPnu0afaWY",
        "_score": 1,
        "_source": {
          "Good_ReferenceNumber": "",
          "..."
        }
      }
    }
  }
}
```

You can also use the above command to inspect the `iaf-demo-anomaly-new` index, where you should find a further 82 items.

As an aside, you can create a composite (albeit complex) command that feeds the url and password directly into the search command:

```bash
curl -k -u elasticsearch-admin:"$(oc extract secret/iaf-system-elasticsearch-es-default-user --to=- --keys=password 2>/dev/null)" https://"$(oc get route iaf-system-es -o=jsonpath='{.spec.host}')"/iafdemo-raw-new/_search
```

> Note: The permitted elasticsearch APIs are controlled by an AllowList in the IBM Automation Foundation. By default, many APIs (such as `count` and `doc`) are not included in this list. Please refer to the [operational datastore section of the IBM Knowledge Centre document on Getting Started with Cloud Paks](https://www-03preprod.ibm.com/support/knowledgecenter/en/cloudpaks_start/cloud-paks/operationaldatastore-cp.html#api-allowlist) for more information on the AllowList.

## Building and extending this repo

For clarity, developer instructions for the code in this repo are moved into a separate [DEVELOPMENT.md](DEVELOPMENT.md) file.
