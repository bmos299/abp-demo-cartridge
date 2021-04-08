module github.ibm.com/automation-base-pak/abp-demo-cartridge

go 1.13

require (
	github.com/Shopify/sarama v1.27.0
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 v2.2.0
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-logr/logr v0.3.0
	github.com/gocarina/gocsv v0.0.0-20200827134620-49f5c3fa2b3e
	github.com/google/uuid v1.2.0
	github.com/minio/minio-go/v7 v7.0.10
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/prometheus/common v0.15.0
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c
	github.ibm.com/automation-base-pak/abp-ai-operator v0.0.0-20210225134141-11775c1d8ffc
	github.ibm.com/automation-base-pak/abp-base-operator v0.0.11-0.20210226052859-d70e8bcd88bc
	github.ibm.com/automation-base-pak/abp-core-operator v0.0.0-20210227034626-6995ee6ac9ca
	github.ibm.com/automation-base-pak/abp-eventprocessing v0.0.97-0.20210226100055-af143b9db6af
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v12.0.0+incompatible
	knative.dev/eventing v0.19.3
	knative.dev/eventing-contrib v0.18.7
	knative.dev/pkg v0.0.0-20210216013737-584933f8280b
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
	k8s.io/api => k8s.io/api v0.19.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.3
	k8s.io/client-go => k8s.io/client-go v0.19.3
)
