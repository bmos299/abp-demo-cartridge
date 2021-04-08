# AI components in demo cartridge

IAF AI Operator depends on AI Foundation components. Currently it uses only the Model Serving components. 
Model Serving is a prerequisite for using AI features in demo cartridge. You would have it installed already if you followed the instructions here [abp-deploy](https://github.ibm.com/automation-base-pak/abp-deploy)

## Demo cartridge data flow

![image](https://raw.github.ibm.com/automation-base-pak/abp-demo-cartridge/main/doc/images/demo-cartridge-dataflow.png)

When the demo cartridge custom resource is created:
  - the demo cartridge operator creates a AI Model custom resource with all the attributes needed to deploy the AI model
  - AI operator, from the attributes specified in the AIModel custom resource, reads the saved model from MINIo and deploys it on Model Serving. This exposes a gRPC endpoint with inference API
  - the demo cartridge operator creates a Flink job which already has the gRPC client java code to invoke the model serving endpoint
  - Flink job reads inconing data from Kafka and invokes the inference API to get the risk assessment value. It adds this value to the data and stores it in ElasticSearch

## AIModel custom resource

This is how the sample AIModel custom resource looks like:

```
apiVersion: ai.automation.ibm.com/v1alpha1
kind: AIModel
metadata:
  name: risk-mapper
spec:
  license:
    accept: true
  name: risk-mapper
  version: 0.0.1
  description: Sample AI Model for Risk Mapping
  content:
    modelType: tensorflow
    path: models/risk-mapper
    bucket: abp-ai
    storageSecret: minio-secret
```

### Note:
For more details on how to deploy your own trained model and include in the cartridge check the AI sections of the [Automation Foundation Playbook](https://pages.github.ibm.com/automation-base-pak/abp-playbook/)