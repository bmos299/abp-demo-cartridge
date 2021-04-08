// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// Licensed Materials - Property of IBM
// 5900-AEO
//
// Copyright IBM Corp. 2020, 2021. All Rights Reserved.
//
// US Government Users Restricted Rights - Use, duplication, or
// disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
// ------------------------------------------------------ {COPYRIGHT-END} ---
/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/common/log"
	basev1beta1 "github.ibm.com/automation-base-pak/abp-base-operator/api/v1beta1"
	epv1alpha1 "github.ibm.com/automation-base-pak/abp-eventprocessing/api/v1alpha1"
	epv1beta1 "github.ibm.com/automation-base-pak/abp-eventprocessing/api/v1beta1"
	epcommon "github.ibm.com/automation-base-pak/abp-eventprocessing/pkg/commoncrd"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

const elasticsearchDemoRawIndexJSON = `{
	"settings":{
		"index":{
			"number_of_shards": 1,
			"number_of_replicas": 0
		}
	},
	"mappings": {
		"properties": {
			"Req_Line_ID": { "type": "keyword", "index": true},
			"Order_Line_ID": { "type": "keyword" },
			"Goods_ID": {  "type": "keyword" },
			"Invoice_ID": { "type": "keyword" },
			"Activity": {  "type": "keyword" },
			"DateTime": { "type": "keyword" },
			"Resource": { "type": "keyword" },
			"Role": { "type": "keyword" },
			"Requisition_Vendor": { "type": "keyword" },
			"Order_Vendor": {  "type": "keyword" },
			"Invoice_Vendor": { "type": "keyword" },
			"Pay_Vendor": { "type": "keyword" },
			"Requisition_Type": { "type": "keyword" },
			"Order_Type": { "type": "keyword" },
			"Purchasing_Group": { "type": "keyword" },
			"Purchasing_Organization": { "type": "keyword" },
			"Material_Group": { "type": "keyword" },
			"Material_Number": { "type": "keyword" },
			"Plant": { "type": "keyword" },
			"Good_ReferenceNumber": {"type": "keyword" },
			"Requisition_Header": { "type": "keyword" },
			"Order_Header": { "type": "keyword" },
			"Invoice_Header": { "type": "keyword" },
			"ClearDoc_Header": { "type": "keyword" },
			"Good_Year": { "type": "keyword" },
			"Invoice_Year": { "type": "keyword" },
			"Order_Line_Amount": { "type": "keyword" },
			"Invoice_Amount": { "type": "float" },
			"Paid_Amount": { "type": "keyword" },
			"Invoice_Document_Date": { "type": "keyword" },
			"Invoice_Due_Date": { "type": "keyword" },
			"Pay_Type": { "type": "keyword" },
			"Pay_Delay": { "type": "integer" },
			"UserType": { "type": "keyword"  },
			"Invoice_Is_Overdue": { "type": "keyword" }
		}
	}
}`

const elasticsearchDemoAnomalyIndexJSON = `{
	"settings":{
		"index":{
			"number_of_shards": 1,
			"number_of_replicas": 0
		}
	},
	"mappings": {
    "properties": {
      "Invoice_ID": { "type": "keyword", "index": true},
      "Invoice_Amount": { "type": "float"},
      "Invoice_Due_Date": { "type": "keyword" },
      "Pay_Type": { "type": "keyword" },
      "Pay_Delay": { "type": "integer" },
      "Risk": { "type": "keyword" }
    }
	}
}`

// Note, as of https://github.ibm.com/automation-base-pak/abp-eventprocessing/pull/68
// the EventProcessor refers to something tangible, typically long-lived, that does the work
// for example a FlinkCluster.
// In contrast, an EventProcessingTask is the unit of work that users submit to the EventProcessor
// and both should be reconciled accordingly.
func (r *IAFDemoReconciler) reconcileEventProcessor(recctx *reconcileContext) error {
	log := r.Log.WithValues("iafdemo", recctx.req.NamespacedName)
	namespace := recctx.iafdemo.Namespace
	licenseAccept := bool(recctx.iafdemo.Spec.License.Accept)

	existingEventProcessingInstance := &epv1beta1.EventProcessor{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: eventProcessorName, Namespace: namespace}, existingEventProcessingInstance)
	if err != nil && errors.IsNotFound(err) {
		log.Info("EventProcessor instance not found. Creating...")

		epInstance := newEventProcessingInstance(r.Cfg.EventProcessorImage, namespace, licenseAccept)

		err = ctrl.SetControllerReference(recctx.iafdemo, epInstance, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}

		err = r.Create(*recctx.ctx, epInstance)
		if err != nil {
			return fmt.Errorf("Failed to create automationbase Event Processing instance: %s", err)
		}
	} else if err != nil {
		return fmt.Errorf("Failed to get automationbase Event Processing instance: %s", err)
	}
	return nil
}

func (r *IAFDemoReconciler) initializeElasticsearchIndices(recctx *reconcileContext) error {
	log := r.Log.WithValues("iafdemo", recctx.req.NamespacedName)

	es, err := r.getElasticsearchInfo(recctx)
	if err != nil {
		return err
	}

	// TODO: when the elastic connector is removed, the `-new` can be removed from these index names.
	log.Info("Creating Elasticsearch index and mapping", "name", eventProcessorInputTopic)
	if err = es.doRequest("PUT", eventProcessorInputTopic+"-new/", elasticsearchDemoRawIndexJSON); err != nil {
		return err
	}

	log.Info("Creating Elasticsearch index and mapping", "name", eventProcessorRiskTopic)
	if err = es.doRequest("PUT", eventProcessorRiskTopic+"-new/", elasticsearchDemoAnomalyIndexJSON); err != nil {
		return err
	}

	return nil
}

type elasticsearchInfo struct {
	endpoint string
	username string
	password string
	cacerts  []byte
}

func configureTransportSecurity(certificate []byte) *tls.Config {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certificate)
	return &tls.Config{RootCAs: certPool}
}

type elasticsearchError struct {
	Error struct {
		RootCause []struct {
			Type      string `json:"type"`
			Reason    string `json:"reason"`
			IndexUUID string `json:"index_uuid"`
			Index     string `json:"index"`
		} `json:"root_cause"`
		Type      string `json:"type"`
		Reason    string `json:"reason"`
		IndexUUID string `json:"index_uuid"`
		Index     string `json:"index"`
	} `json:"error"`
	Status int `json:"status"`
}

func (es elasticsearchInfo) doRequest(method, path, body string) error {
	transportConfig := &tls.Config{}
	if es.cacerts != nil {
		transportConfig = configureTransportSecurity(es.cacerts)
	}
	client := &http.Client{
		Timeout: time.Second * 10, // The default http.Client has no timeout, so it is good to specify one
		Transport: &http.Transport{
			TLSClientConfig: transportConfig,
		},
	}

	url := fmt.Sprintf("%s/%s", es.endpoint, path)
	if !strings.HasPrefix(url, "http") { // Only add the protocol if it's not already present.
		if len(es.cacerts) > 0 { // Use https if a CA cert is present
			url = fmt.Sprintf("https://%s", url)
		} else {
			url = fmt.Sprintf("http://%s", url)
		}
	}
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return err
	}

	req.SetBasicAuth(es.username, es.password)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, _ := ioutil.ReadAll(resp.Body) // Read to EOF so transport can be re-used
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Unpack the response, and check if it only failed because it already existed.
		var esError elasticsearchError
		if err = json.Unmarshal(respBody, &esError); err != nil {
			return err
		}
		if esError.Error.Type == "resource_already_exists_exception" {
			return nil
		}
		return fmt.Errorf("Elasticsearch request %v %v failed, status %v", method, path, resp.StatusCode)
	}
	return nil
}

func (r *IAFDemoReconciler) getElasticsearchInfo(recctx *reconcileContext) (elasticsearchInfo, error) {
	es := elasticsearchInfo{}
	namespace := recctx.iafdemo.Namespace

	// Get the CartridgeRequirements to extract some information from its status
	reqInst := &basev1beta1.CartridgeRequirements{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: iafCartridgeReqInstanceName, Namespace: namespace}, reqInst)
	if err != nil && errors.IsNotFound(err) {
		err = fmt.Errorf("Failed to find CartridgeRequirements %s in Namespace %s: %w", iafCartridgeReqInstanceName, namespace, err)
		return es, err
	} else if err != nil {
		return es, err
	}

	if reqInst.Status.Components == nil || reqInst.Status.Components.ElasticSearch == nil {
		return es, fmt.Errorf("Could not get ElasticSearch information from CartridgeRequirements %s in Namespace %s", iafCartridgeReqInstanceName, namespace)
	}

	elasticsearchAuthSecretName := ""
	elasticsearchCertsSecretName := ""
	for _, esEndpoint := range reqInst.Status.Components.ElasticSearch.Endpoints {
		if esEndpoint.Scope == basev1beta1.EndpointScopeInternal {
			// Get the Internal endpoint, not the External (or any other) endpoint.
			es.endpoint = esEndpoint.URI
			if esEndpoint.Authentication != nil && esEndpoint.Authentication.Secret != nil {
				elasticsearchAuthSecretName = esEndpoint.Authentication.Secret.SecretName
			}
			if esEndpoint.CASecret != nil {
				elasticsearchCertsSecretName = esEndpoint.CASecret.SecretName
			}
		}
	}
	if len(es.endpoint) == 0 {
		return es, fmt.Errorf("Could not get ElasticSearch Internal endpoint from CartridgeRequirements %s in Namespace %s", iafCartridgeReqInstanceName, namespace)
	}

	elasticsearchAuthSecret := &corev1.Secret{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: elasticsearchAuthSecretName, Namespace: namespace}, elasticsearchAuthSecret)
	if err != nil && errors.IsNotFound(err) {
		err = fmt.Errorf("Failed to find Secret %s in Namespace %s: %w", elasticsearchAuthSecretName, namespace, err)
		return es, err
	} else if err != nil {
		return es, err
	}

	username, ok := elasticsearchAuthSecret.Data["username"]
	if !ok {
		err = fmt.Errorf("Failed to get key 'username' in Secret %s in Namespace %s: %w", elasticsearchAuthSecretName, namespace, err)
		return es, err
	}
	es.username = string(username)

	password, ok := elasticsearchAuthSecret.Data["password"]
	if !ok {
		err = fmt.Errorf("Failed to get key 'password' in Secret %s in Namespace %s: %w", elasticsearchAuthSecretName, namespace, err)
		return es, err
	}
	es.password = string(password)

	if elasticsearchCertsSecretName != "" {
		elasticsearchCertsSecret := &corev1.Secret{}
		err = r.Get(*recctx.ctx, types.NamespacedName{Name: elasticsearchCertsSecretName, Namespace: namespace}, elasticsearchCertsSecret)
		if err != nil && errors.IsNotFound(err) {
			err = fmt.Errorf("Failed to find Secret %s in Namespace %s: %w", elasticsearchCertsSecretName, namespace, err)
			return es, err
		} else if err != nil {
			return es, err
		}

		cacerts, ok := elasticsearchCertsSecret.Data["ca.crt"]
		if !ok {
			err = fmt.Errorf("Failed to get key 'ca.crt' in Secret %s in Namespace %s: %w", elasticsearchCertsSecretName, namespace, err)
			return es, err
		}
		es.cacerts = cacerts
	}

	return es, nil
}

func (r *IAFDemoReconciler) reconcileEventProcessingTask(recctx *reconcileContext) error {
	log := r.Log.WithValues("iafdemo", recctx.req.NamespacedName)
	namespace := recctx.iafdemo.Namespace
	licenseAccept := bool(recctx.iafdemo.Spec.License.Accept)

	existingEventProcessingTaskInstance := &epv1alpha1.EventProcessingTask{}

	err := r.Get(*recctx.ctx, types.NamespacedName{Name: eventProcessingTaskInstanceName, Namespace: namespace}, existingEventProcessingTaskInstance)

	predictorEndPoint, _ := r.getPredictorURL(recctx)
	log.Info("predictorEndPoint: " + predictorEndPoint)

	if err != nil && errors.IsNotFound(err) {
		log.Info("EventProcessingTask instance not found. Creating...")

		epTaskInstance := newEventProcessingTaskInstance(r.Cfg.EventProcessingTaskImage, namespace, licenseAccept, predictorEndPoint)

		err = ctrl.SetControllerReference(recctx.iafdemo, epTaskInstance, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}

		err = r.Create(*recctx.ctx, epTaskInstance)
		if err != nil {
			return fmt.Errorf("Failed to create automationbase EventProcessingTask instance: %s", err)
		}
	} else if err != nil {
		return fmt.Errorf("Failed to get automationbase EventProcessingTask instance: %s", err)
	}
	return nil
}

func newEventProcessingInstance(eventProcessorImage, namespace string, licenseAccept bool) *epv1beta1.EventProcessor {
	saToUse := eventProcessorServiceAccountName
	// Use the backtick here so we can have a formatted string with quotes etc
	// this is useful so one can get the job logs from the JobManager pod and in the UI
	logConfigAsMap := map[string]string{
		"log4j-console.properties": `rootLogger.level = INFO
		rootLogger.appenderRef.file.ref = LogFile
		rootLogger.appenderRef.console.ref = LogConsole
		appender.file.name = LogFile
		appender.file.type = File
		appender.file.append = false
		appender.file.fileName = ${sys:log.file}
		appender.file.layout.type = PatternLayout
		appender.file.layout.pattern = %d{yyyy-MM-dd HH:mm:ss,SSS} %-5p %-60c %x - %m%n
		appender.console.name = LogConsole
		appender.console.type = CONSOLE
		appender.console.layout.type = PatternLayout
		appender.console.layout.pattern = %d{yyyy-MM-dd HH:mm:ss,SSS} %-5p %-60c %x - %m%n
		logger.akka.name = akka
		logger.akka.level = INFO
		logger.kafka.name= org.apache.kafka
		logger.kafka.level = INFO
		logger.hadoop.name = org.apache.hadoop
		logger.hadoop.level = INFO
		logger.zookeeper.name = org.apache.zookeeper
		logger.zookeeper.level = INFO
		logger.netty.name = org.apache.flink.shaded.akka.org.jboss.netty.channel.DefaultChannelPipeline
		logger.netty.level = OFF`,
		"logback-console.xml": `<configuration>
		<appender name="console" class="ch.qos.logback.core.ConsoleAppender">
			<encoder>
				<pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{60} %X{sourceThread} - %msg%n</pattern>
			</encoder>
		</appender>
		<appender name="file" class="ch.qos.logback.core.FileAppender">
			<file>${log.file}</file>
			<append>false</append>
			<encoder>
				<pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{60} %X{sourceThread} - %msg%n</pattern>
			</encoder>
		</appender>
		<root level="INFO">
			<appender-ref ref="console"/>
			<appender-ref ref="file"/>
		</root>
		<logger name="akka" level="INFO" />
		<logger name="org.apache.kafka" level="INFO" />
		<logger name="org.apache.hadoop" level="INFO" />
		<logger name="org.apache.zookeeper" level="INFO" />
		<logger name="org.apache.flink.shaded.akka.org.jboss.netty.channel.DefaultChannelPipeline" level="ERROR" />
	</configuration>`,
	}
	return &epv1beta1.EventProcessor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      eventProcessorName,
			Namespace: namespace,
			Annotations: map[string]string{
				"com.ibm.automation.cartridge": iafCartridgeInstanceName,
			},
		},
		Spec: epv1beta1.EventProcessorSpec{
			License: epcommon.License{
				Accept: licenseAccept,
			},
			Version: "1.0.0",
			Flink: &epv1beta1.FlinkSpec{
				// On - using self-signed certs
				TLS: &epv1beta1.TLS{},
				// Empty option so we get an auto-generated user+pass combination
				// that the script inside the Flink image can use (referencing their locations)
				Authentication:     &epv1beta1.Authentication{},
				ServiceAccountName: &saToUse,
				Image:              eventProcessorImage,
				TaskManager: &epv1beta1.TaskManagerSpec{
					Replicas: int32(1),
				},
				LogConfig: logConfigAsMap,
				Properties: map[string]string{
					"jobmanager.memory.flink.size":    "600mb",
					"jobmanager.memory.process.size":  "1200mb",
					"taskmanager.memory.flink.size":   "600mb",
					"taskmanager.memory.process.size": "1200mb",
				},
			},
		},
	}
}

func newEventProcessingTaskInstance(image, namespace string, licenseAccept bool, predictorEndPoint string) *epv1alpha1.EventProcessingTask {
	saToUse := eventProcessorServiceAccountName
	programArgs := "?program-args=--groupId " + eventProcessorGroup + " --rawTopic " + eventProcessorInputTopic + " --riskTopic " + eventProcessorRiskTopic + " --esRawIndex " + eventProcessorInputTopic + " --esRiskIndex " + eventProcessorRiskTopic

	if len(predictorEndPoint) > 0 {
		log.Info("Added model predictor URL to the job")
		programArgs += " --modelPredictorURL " + predictorEndPoint
	}
	//programArgs += "\""

	eventProcessorArgsToUse := []string{"-j", "/opt/flink/demo/demo-flink-job-1.0-SNAPSHOT.jar", "-q", programArgs}

	log.Info("Job args for this EventProcessingTask to use is: ", fmt.Sprintf("%v", eventProcessorArgsToUse))

	return &epv1alpha1.EventProcessingTask{
		ObjectMeta: metav1.ObjectMeta{
			Name:      eventProcessingTaskInstanceName,
			Namespace: namespace,
			Annotations: map[string]string{
				"com.ibm.automation.cartridge": iafCartridgeInstanceName,
			},
		},
		Spec: epv1alpha1.EventProcessingTaskSpec{
			License: epcommon.License{
				Accept: licenseAccept,
			},
			Version:            "1.0.0",
			ProcessorName:      eventProcessorName,
			ServiceAccountName: &saToUse,
			Image:              image,
			RestartPolicy:      "OnFailure",
			Command:            []string{"/curl-helper.sh"},
			Args:               eventProcessorArgsToUse,
		},
	}
}
