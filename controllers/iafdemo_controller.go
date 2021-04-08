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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	basev1beta1 "github.ibm.com/automation-base-pak/abp-base-operator/api/v1beta1"
	kafkav1beta1 "github.ibm.com/automation-base-pak/abp-base-operator/api/v1beta1kafka"
	corev1beta1 "github.ibm.com/automation-base-pak/abp-core-operator/api/v1beta1"
	democartridgev1 "github.ibm.com/automation-base-pak/abp-demo-cartridge/api/v1"
	"github.ibm.com/automation-base-pak/abp-demo-cartridge/pkg/config"
)

const (
	deployedServerName              = "demoserver"
	deployedProducerName            = "demoproducer"
	iafCartridgeInstanceName        = "iafdemo"
	iafCartridgeReqInstanceName     = "iaf-cartridgerequirements-instance"
	automationBaseInstanceName      = "iaf-automationbase-instance"
	eventProcessorName              = "iaf-eventprocessor-instance"
	eventProcessingTaskInstanceName = "iaf-eventprocessing-task-instance"
	eventProcessorInputTopic        = iafCartridgeInstanceName + "-raw"
	eventProcessorRiskTopic         = iafCartridgeInstanceName + "-anomaly"
	eventProcessorGroup             = iafCartridgeInstanceName + "-flink-processor"

	eventStreamInstance = "iaf-eventstream-"

	// Real request would look like this
	// curl --cacert /my/path/to/tls.crt -u eventprocessing-admin:thepassword -X POST -H 'Content-Type: application/json' https://localhost:8081/jars/fb55c3b0-19d0-4f52-8f78-b248bf69efbc_demo-flink-job-1.0-SNAPSHOT.jar/run?program-args=--groupId%20iafdemo-flink-processor%20--rawTopic%20iafdemo-raw%20--riskTopic%20iafdemo-anomaly%20--esRawIndex%20iafdemo-raw%20--esRiskIndex%20iafdemo-anomaly

	eventProcessorServiceAccountName = "iaf-eventprocessing-operator-default"
)

// IAFDemoReconciler reconciles a IAFDemo object
type IAFDemoReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Cfg    *config.Config
}

type reconcileContext struct {
	ctx     *context.Context
	req     *ctrl.Request
	iafdemo *democartridgev1.IAFDemo
}

// +kubebuilder:rbac:groups=democartridge.ibm.com,resources=iafdemoes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=democartridge.ibm.com,resources=iafdemoes/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=democartridge.ibm.com,resources=iafdemoes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.automation.ibm.com,resources=cartridges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=base.automation.ibm.com,resources=automationbases,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=ibmevents.ibm.com,resources=kafkatopics,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=base.automation.ibm.com,resources=cartridgerequirements,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventprocessing.automation.ibm.com,resources=eventprocessors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventprocessing.automation.ibm.com,resources=eventprocessingtasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=automation.ibm.com,resources=eventprocessors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ai.automation.ibm.com,resources=aimodels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sources.knative.dev,resources=kafkasources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=messaging.knative.dev,resources=inmemorychannels;subscriptions,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=ai.automation.ibm.com,resources=airuntimes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ai.automation.ibm.com,resources=airuntimes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ai.automation.ibm.com,resources=aideployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ai.automation.ibm.com,resources=aideployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ai.automation.ibm.com,resources=aimodels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ai.automation.ibm.com,resources=aimodels/status,verbs=get;update;patch

func (r *IAFDemoReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var retryWaitTime = 30 * time.Second
	ctx := context.Background()
	log := r.Log.WithValues("iafdemo", req.NamespacedName)
	iafdemo := &democartridgev1.IAFDemo{}
	err := r.Get(ctx, req.NamespacedName, iafdemo)
	if err != nil && errors.IsNotFound(err) {
		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		// Return and don't requeue
		log.Info("automationbase resource not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// IAFDemo CR found, proceeding with steps to reconcile.
	// creating a reconcile context to keep state for stages of reconcile loop.
	recctx := &reconcileContext{
		ctx:     &ctx,
		req:     &req,
		iafdemo: iafdemo,
	}

	// handle cartridge CR
	//log.Info("Creating the Cartridge CR and Issuing it")
	retryAfter, err := r.createCartridge(recctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if retryAfter {
		log.Info("Waiting for the Cartridge registration to complete")
		return ctrl.Result{RequeueAfter: retryWaitTime}, nil
	}

	retryAfter, err = r.createAutomationBase(recctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if retryAfter {
		log.Info("Waiting for the AutomationBase installation to complete")
		return ctrl.Result{RequeueAfter: retryWaitTime}, nil
	}

	retryAfter, err = r.createCartridgeRequirements(recctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if retryAfter {
		log.Info("Waiting for the CartridgeRequirements registration to complete")
		return ctrl.Result{RequeueAfter: retryWaitTime}, nil
	}

	retryAfter, err = r.setupAIModels(recctx)
	if err != nil {
		log.Error(err, "AI is not configured")
	}
	if retryAfter {
		log.Info("Waiting for the setting up of AIModels to complete")
		return ctrl.Result{RequeueAfter: retryWaitTime}, nil
	}

	retryAfter, err = r.createEventStream(recctx, eventProcessorInputTopic) // demo raw
	if err != nil {
		return ctrl.Result{}, err
	}
	if retryAfter {
		log.Info("Waiting for the EventStream registration to complete")
		return ctrl.Result{RequeueAfter: retryWaitTime}, nil
	}

	retryAfter, err = r.createEventStream(recctx, eventProcessorRiskTopic) // demo anomaly
	if err != nil {
		return ctrl.Result{}, err
	}
	if retryAfter {
		log.Info("Waiting for the EventStream registration to complete")
		return ctrl.Result{RequeueAfter: retryWaitTime}, nil
	}

	err = r.reconcileEventProcessor(recctx)
	if err != nil {
		log.Error(err, "Failed to reconcile Event Processing CR")
		return ctrl.Result{}, err
	}

	err = r.initializeElasticsearchIndices(recctx)
	if err != nil {
		log.Error(err, "Failed to initialize Elasticsearch Indices")
		return ctrl.Result{}, err
	}

	err = r.reconcileEventProcessingTask(recctx)
	if err != nil {
		log.Error(err, "Failed to reconcile Event Processing Task CR")
		return ctrl.Result{}, err
	}

	// The Knative piece can be enabled/disabled via an environment variable
	if strings.EqualFold(r.Cfg.UseKnative, "true") {
		err = r.reconcileKnative(recctx, deployedServerName)
		if err != nil {
			log.Error(err, "Failed to create Knative connection")
			return ctrl.Result{}, err
		}
	}

	// create producer microservice M1
	err = r.reconcileMicroservice(recctx, deployedProducerName)
	if err != nil {
		return ctrl.Result{}, err
	}

	// create server microservice M2
	err = r.reconcileMicroservice(recctx, deployedServerName)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Reconcile at end returning nil err or SUCESS!!")
	return ctrl.Result{}, nil
}

func (r *IAFDemoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&democartridgev1.IAFDemo{}).
		Complete(r)
}

func (r *IAFDemoReconciler) createCartridge(recctx *reconcileContext) (bool, error) {
	namespace := recctx.iafdemo.Namespace
	err := r.createCartridgeInstance(recctx)
	if err != nil {
		return false, err
	}

	existingIafCartridgeInstance := &corev1beta1.Cartridge{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: iafCartridgeInstanceName, Namespace: namespace}, existingIafCartridgeInstance)
	if err != nil {
		return false, err
	}

	runningCondition := existingIafCartridgeInstance.Status.Conditions.GetCondition("Ready")
	if runningCondition != nil && runningCondition.Status == corev1.ConditionTrue {
		return false, nil
	} else {
		return true, nil
	}
}

func (r *IAFDemoReconciler) createCartridgeInstance(recctx *reconcileContext) error {
	namespace := recctx.iafdemo.Namespace
	licenseAccept := bool(recctx.iafdemo.Spec.License.Accept)

	existingCartridgeInstance := &corev1beta1.Cartridge{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: iafCartridgeInstanceName, Namespace: namespace}, existingCartridgeInstance)
	if err != nil && errors.IsNotFound(err) {
		cartridgeInstance := newCartridgeInstance(namespace, licenseAccept)
		err = ctrl.SetControllerReference(recctx.iafdemo, cartridgeInstance, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}
		err = r.Create(*recctx.ctx, cartridgeInstance)
		if err != nil {
			return fmt.Errorf("Failed to create automationbase cartridge instance: %s", err)
		}
	} else if err != nil {
		return fmt.Errorf("Failed to get automationbase cartridge instance: %s", err)
	}
	return nil
}

func newCartridgeInstance(namespace string, licenseAccept bool) *corev1beta1.Cartridge {
	return &corev1beta1.Cartridge{
		ObjectMeta: metav1.ObjectMeta{
			Name:      iafCartridgeInstanceName,
			Namespace: namespace,
		},
		Spec: corev1beta1.CartridgeSpec{
			Version: "1.0.0",
			License: corev1beta1.License{
				Accept: licenseAccept,
			},
		},
	}
}

func (r *IAFDemoReconciler) createAutomationBase(recctx *reconcileContext) (bool, error) {
	namespace := recctx.iafdemo.Namespace
	err := r.createAutomationBaseInstance(recctx)
	if err != nil {
		return false, err
	}

	existingAutomationBaseInstance := &basev1beta1.AutomationBase{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: automationBaseInstanceName, Namespace: namespace}, existingAutomationBaseInstance)
	if err != nil {
		return false, err
	}

	runningCondition := existingAutomationBaseInstance.Status.Conditions.GetCondition("Ready")
	if runningCondition != nil && runningCondition.Status == corev1.ConditionTrue {
		return false, nil
	} else {
		return true, nil
	}
}

func (r *IAFDemoReconciler) createAutomationBaseInstance(recctx *reconcileContext) error {
	namespace := recctx.iafdemo.Namespace
	licenseAccept := bool(recctx.iafdemo.Spec.License.Accept)
	existingAutomationBaseInstance := &basev1beta1.AutomationBase{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: automationBaseInstanceName, Namespace: namespace}, existingAutomationBaseInstance)
	if err != nil && errors.IsNotFound(err) {
		automationBaseInstance := newAutomationBaseInstance(namespace, licenseAccept)
		err = ctrl.SetControllerReference(recctx.iafdemo, automationBaseInstance, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}

		err = r.Create(*recctx.ctx, automationBaseInstance)
		if err != nil {
			return fmt.Errorf("Failed to create AutomationBase instance: %s", err)
		}
	} else if err != nil {
		return fmt.Errorf("Failed to get AutomationBase instance: %s", err)
	}
	return nil
}

func newAutomationBaseInstance(namespace string, licenseAccept bool) *basev1beta1.AutomationBase {
	return &basev1beta1.AutomationBase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      automationBaseInstanceName,
			Namespace: namespace,
		},
		Spec: basev1beta1.AutomationBaseSpec{
			Version:       "1.0.0",
			Kafka:         &kafkav1beta1.KafkaSpec{},
			ElasticSearch: &basev1beta1.AutomationBaseElasticsearchSpec{},
			License: basev1beta1.License{
				Accept: licenseAccept,
			},
			TLS: &basev1beta1.TLS{},
		},
	}
}

func (r *IAFDemoReconciler) createCartridgeRequirements(recctx *reconcileContext) (bool, error) {
	namespace := recctx.iafdemo.Namespace
	err := r.createCartridgeRequirementsInstance(recctx)
	if err != nil {
		return false, err
	}

	existingCartridgeReqInstance := &basev1beta1.CartridgeRequirements{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: iafCartridgeReqInstanceName, Namespace: namespace}, existingCartridgeReqInstance)
	if err != nil {
		return false, err
	}

	runningCondition := existingCartridgeReqInstance.Status.Conditions.GetCondition("Ready")
	if runningCondition != nil && runningCondition.Status == corev1.ConditionTrue {
		return false, nil
	} else {
		return true, nil
	}
}

func (r *IAFDemoReconciler) createCartridgeRequirementsInstance(recctx *reconcileContext) error {
	namespace := recctx.iafdemo.Namespace
	licenseAccept := bool(recctx.iafdemo.Spec.License.Accept)
	existingCartridgeReqInstance := &basev1beta1.CartridgeRequirements{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: iafCartridgeReqInstanceName, Namespace: namespace}, existingCartridgeReqInstance)
	if err != nil && errors.IsNotFound(err) {
		cartridgeReqInstance := newCartridgeRequirementsInstance(namespace, licenseAccept)
		err = ctrl.SetControllerReference(recctx.iafdemo, cartridgeReqInstance, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}
		err = r.Create(*recctx.ctx, cartridgeReqInstance)
		if err != nil {
			return fmt.Errorf("Failed to create cartridge requirements instance: %s", err)
		}
	} else if err != nil {
		return fmt.Errorf("Failed to get cartridge requirements instance: %s", err)
	}
	return nil
}

func newCartridgeRequirementsInstance(namespace string, licenseAccept bool) *basev1beta1.CartridgeRequirements {
	return &basev1beta1.CartridgeRequirements{
		ObjectMeta: metav1.ObjectMeta{
			Name:      iafCartridgeReqInstanceName,
			Namespace: namespace,
			Annotations: map[string]string{
				"com.ibm.automation.cartridge": iafCartridgeInstanceName,
			},
		},
		Spec: basev1beta1.CartridgeRequirementsSpec{
			License: basev1beta1.License{
				Accept: licenseAccept,
			},
			Version: "1.0.0",
			Requirements: []basev1beta1.RequirementType{
				basev1beta1.EventProcessors,
				basev1beta1.Events,
			},
		},
	}
}
