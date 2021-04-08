When an ABPDemo resource is created, the demo-cartridge operator automatically creates a deployment, service, and a route using the very simple server here.

To connect events from the 'demo-anomolies' topic in Kafka to this server, we can use Knative. First, Knative must be installed in the cluster. In OpenShift, the serverless-operator can be added to the cluster (via the OperatorHub, for example), and then KnativeServing and KnativeEventing instances can be created, for example with these resources:
```bash
oc apply -f - <<EOF
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
EOF
```

To get events from Kafka into a Knative Channel, the KafkaSource CRD needs to be installed:
```bash
oc apply -f "https://github.com/knative/eventing-contrib/releases/download/v0.17.0/kafka-source.yaml"
```

Finally, the resources connecting the topic to the demo server deployment can be created:
```bash
KAFKA_CLUSTER_ADDRESS="sandbox-kafka-bootstrap.kafka:9092"
oc apply -f - <<EOF
apiVersion: sources.knative.dev/v1alpha1
kind: KafkaSource
metadata:
  name: demo-anomolies
spec:
  consumerGroup: knative-group
  bootstrapServers:
    - $KAFKA_CLUSTER_ADDRESS
  topics:
    - demo-anomolies
  sink:
    ref:
      apiVersion: messaging.knative.dev/v1alpha1
      kind: InMemoryChannel
      name: demo-anomolies
---
apiVersion: messaging.knative.dev/v1alpha1
kind: InMemoryChannel
metadata:
  name: demo-anomolies
---
apiVersion: messaging.knative.dev/v1beta1
kind: Subscription
metadata:
  name: echo-on-events
spec:
  channel:
    apiVersion: messaging.knative.dev/v1beta1
    kind: InMemoryChannel
    name: demo-anomolies
  subscriber:
    ref:
      apiVersion: v1
      kind: Service
      name: demoserver
EOF
```
