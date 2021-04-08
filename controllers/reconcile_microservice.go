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
	"fmt"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	basev1beta1 "github.ibm.com/automation-base-pak/abp-base-operator/api/v1beta1"
	"github.ibm.com/automation-base-pak/abp-demo-cartridge/pkg/common"
	"github.ibm.com/automation-base-pak/abp-demo-cartridge/pkg/server"
)

func (r *IAFDemoReconciler) reconcileMicroservice(recctx *reconcileContext, deployedName string) error {
	namespace := recctx.iafdemo.Namespace
	messagesPerGroup := recctx.iafdemo.Spec.MessagesPerGroup
	secondsToPause := recctx.iafdemo.Spec.SecondsToPause
	sequenceRepititions := recctx.iafdemo.Spec.SequenceRepititions

	log := r.Log.WithValues("iafdemo", recctx.req.NamespacedName)

	// Get the CartridgeRequirements to extract some information from its status
	cartridgeReqInstance := &basev1beta1.CartridgeRequirements{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: iafCartridgeReqInstanceName, Namespace: namespace}, cartridgeReqInstance)
	if err != nil && errors.IsNotFound(err) {
		return fmt.Errorf("Failed to find CartridgeRequirements %s in Namespace %s: %w", iafCartridgeReqInstanceName, namespace, err)
	} else if err != nil {
		return err
	}

	shortName := strings.Replace(deployedName, "demo", "", 1) // strip of demo leaving "server" or "producer"
	log.Info("shortName is " + shortName)

	envVars := []corev1.EnvVar{{
		Name:  "FUNCTION",
		Value: shortName,
	}, {
		Name:  "KAFKA_TOPIC",
		Value: eventProcessorInputTopic,
	}, {
		Name:  "MESSAGES_PER_GROUP",
		Value: messagesPerGroup,
	}, {
		Name:  "SECONDS_TO_PAUSE",
		Value: secondsToPause,
	}, {
		Name:  "SAMPLE_REPITITIONS",
		Value: sequenceRepititions,
	}}

	internalTLSEndpointFound := false
	if cartridgeReqInstance.Status.Components == nil || cartridgeReqInstance.Status.Components.Kafka == nil {
		return fmt.Errorf("Failed to get KafkaBootstrap servers from CartridgeRequirements %s in Namespace %s", iafCartridgeReqInstanceName, namespace)
	}
	for _, endpoint := range cartridgeReqInstance.Status.Components.Kafka.Endpoints {
		if endpoint.Name == "internal-service-tls" {
			envVars = append(envVars, corev1.EnvVar{
				Name:  "BOOTSTRAP_SERVERS",
				Value: endpoint.BootstrapServers,
			})

			if len(endpoint.CASecret.SecretName) > 0 {
				log.Info("Kafka CA secret found in CartridgeRequirements")
				envVars = append(envVars, corev1.EnvVar{
					Name: "KAFKA_CA_CERT_PEM",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: endpoint.CASecret.SecretName,
							},
							Key: endpoint.CASecret.Key,
						},
					},
				})
			}

			if len(endpoint.Authentication.Secret.SecretName) > 0 {
				log.Info("Kafka Authentication secret found in CartridgeRequirements")
				envVars = append(envVars, corev1.EnvVar{
					Name:  "KAFKA_USERNAME",
					Value: endpoint.Authentication.Secret.SecretName,
				}, corev1.EnvVar{
					Name: "KAFKA_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: endpoint.Authentication.Secret.SecretName,
							},
							Key: "password",
						},
					},
				})
			}

			internalTLSEndpointFound = true
			break
		}
	}
	if !internalTLSEndpointFound {
		return fmt.Errorf("Failed to find internal-service-tls Kafka endpoint in CartridgeRequirements %s in Namespace %s", iafCartridgeReqInstanceName, namespace)
	}

	// Create deployment if not present
	mDeployment := &appsv1.Deployment{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: deployedName, Namespace: namespace}, mDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Microservice " + deployedName + " deployment not found. Creating...")
		replicas := int32(1)
		mDeployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        deployedName,
				Namespace:   namespace,
				Annotations: map[string]string{},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: common.Labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: common.Labels,
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: "iaf-demo-cartridge-operator",
						Containers: []corev1.Container{{
							Image:           r.Cfg.ServerImage,
							Name:            shortName,
							Env:             envVars,
							ImagePullPolicy: corev1.PullAlways,
						}},
					},
				},
			},
		}
		err = ctrl.SetControllerReference(recctx.iafdemo, mDeployment, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}

		err = r.Create(*recctx.ctx, mDeployment)
		if err != nil {
			return fmt.Errorf("Failed to create new Deployment %s in Namespace %s: %s", mDeployment.Name, mDeployment.Namespace, err)
		}
	} else if err != nil {
		return err
	}

	// Create service if not present
	mService := &corev1.Service{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: deployedName, Namespace: namespace}, mService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Microservice " + deployedName + " service not found. Creating...")
		mService = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        deployedName,
				Namespace:   recctx.iafdemo.Namespace,
				Annotations: map[string]string{},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(server.Port),
					Protocol:   "TCP",
				}},
				Selector: common.Labels,
				Type:     corev1.ServiceTypeClusterIP,
			},
		}
		err = ctrl.SetControllerReference(recctx.iafdemo, mService, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}

		err = r.Create(*recctx.ctx, mService)
		if err != nil {
			return fmt.Errorf("Failed to create new Service %s in Namespace %s: %s", mService.Name, mService.Namespace, err)
		}
	} else if err != nil {
		return err
	}

	// Create Route if not present
	mRoute := &routev1.Route{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: deployedName, Namespace: recctx.iafdemo.Namespace}, mRoute)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Microservice " + deployedName + " route not found. Creating...")
		mRoute = &routev1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:        deployedName,
				Namespace:   recctx.iafdemo.Namespace,
				Annotations: map[string]string{},
			},
			Spec: routev1.RouteSpec{
				Port: &routev1.RoutePort{
					TargetPort: intstr.FromInt(server.Port),
				},
				To: routev1.RouteTargetReference{
					Kind: "Service",
					Name: deployedName,
				},
			},
		}
		err = ctrl.SetControllerReference(recctx.iafdemo, mRoute, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}

		err = r.Create(*recctx.ctx, mRoute)
		if err != nil {
			return fmt.Errorf("Failed to create new Route %s in Namespace %s: %s", mRoute.Name, mRoute.Namespace, err)
		}
	} else if err != nil {
		return err
	}
	return nil
}
