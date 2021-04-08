# demo-flink-processor

Flink based event processor for demoing end to end ABP flow for end of Sprint 6.

This  will start a Flink cluster and submit a Flink job on that cluster. The Flink job will:

- read JSON messages from Kafka topic `iafdemo-raw` representing orders and invoices.
- process all delayed payments and assign a risk level to it based on amount and how delayed it is.
- send risk messsages to Kafka topic `iafdemo-anomaly`. 

# Instructions for building the demo flink processor image:

1. Build the demo Flink processor image: 
    ```
    docker build . -t us.icr.io/abp-scratchpad/iaf-demo-flink-processor:<yourtag>
    ```

1. Push the image to a container registry that can be accessed from the OpenShift cluster, for example with:

```
  docker push us.icr.io/abp-scratchpad/iaf-demo-flink-processor:<yourtag>
```

1. You should now be able to specify this image in `EventProcessor` and `EventProcessingTask` custom resources (modify either config.go or the reconciler for event processing resources, such that your image is explicitly used). Please see the playbook for the exact specification of these resources.


# Instructions for building and running demo job locally (for testing Flink job logic):

1. Start Kafka on your machine. 
    ```
    docker run -it -p 2181:2181 -p 9092:9092 -e "KAFKA_ADVERTISED_LISTNERS=localhost" paspaola/kafka-one-container
    ```

1. Install and start the Flink cluster locally: https://ci.apache.org/projects/flink/flink-docs-release-1.11/getting-started/tutorials/local_setup.html

    ```
    cd <path-to-flink-install>
    ./bin/start-cluster.sh
    ```

1. Build and run the demo Flink job: 

    ```
    cd <path-to-this-dir>
    mvn clean install
    flink run target/demo-flink-job-1.0-SNAPSHOT.jar
    ```

1. Start the demo producer locally with instructions from: https://github.ibm.com/automation-base-pak/abp-demo-cartridge
    ```
    cd <path-to-abp-demo-cartridge>
    FUNCTION=producer BOOTSTRAP_SERVERS=localhost:9092 ./manager
    ```

1. Check that the `risk` messages are appearing on Kafka topics and in the TaskManager's STDOUT.
    ```
    kafkacat -b localhost:9092 -C -t iafdemo-anomaly
    ```

1. You should be able to see the job logs on the console or on the Flink Web UI.