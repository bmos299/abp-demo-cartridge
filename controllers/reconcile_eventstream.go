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

// TO DO: Change variable / method names to remove EventStream; replace with Kafka (Ex: replace newEventStreamInstance with new KafkaTopicInstance)
package controllers

import (
	"fmt"

	kafkatopics "github.ibm.com/automation-base-pak/abp-base-operator/api/v1beta1kafka"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *IAFDemoReconciler) createEventStream(recctx *reconcileContext, topicName string) (bool, error) {
	namespace := recctx.iafdemo.Namespace
	err := r.reconcileEventStream(recctx, topicName)
	if err != nil {
		return false, err
	}

	esInstanceName := topicName
	existingEventStreamInstance := &kafkatopics.KafkaTopic{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: esInstanceName, Namespace: namespace}, existingEventStreamInstance)
	if err != nil {
		return false, err
	} else {
		return false, nil
	}
}

func (r *IAFDemoReconciler) reconcileEventStream(recctx *reconcileContext, topicName string) error {
	log := r.Log.WithValues("iafdemo", recctx.req.NamespacedName)
	namespace := recctx.iafdemo.Namespace
	licenseAccept := bool(recctx.iafdemo.Spec.License.Accept)
	esInstanceName := topicName
	existingEventStreamInstance := &kafkatopics.KafkaTopic{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: esInstanceName, Namespace: namespace}, existingEventStreamInstance)
	if err != nil && errors.IsNotFound(err) {
		log.Info("EventStream instance not found. Creating...")

		esInstance := newEventStreamInstance(namespace, topicName, licenseAccept)

		err = ctrl.SetControllerReference(recctx.iafdemo, esInstance, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}

		err = r.Create(*recctx.ctx, esInstance)
		if err != nil {
			return fmt.Errorf("Failed to create event stream instance: %s", err)
		}
	} else if err != nil {
		return fmt.Errorf("Failed to get event stream instance: %s", err)
	}

	return nil
}

func newEventStreamInstance(namespace string, topicName string, licenseAccept bool) *kafkatopics.KafkaTopic {
	labels := map[string]string{
		"ibmevents.ibm.com/cluster": "iaf-system",
	}

	return &kafkatopics.KafkaTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KafkaTopic",
			APIVersion: "ibmevents.ibm.com/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      topicName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: kafkatopics.KafkaTopicSpec{
			Partitions: 1,
			Replicas:   1,
			Config: map[string]intstr.IntOrString{
				"retention.ms":  intstr.FromInt(604800000),
				"segment.bytes": intstr.FromInt(1073741824),
			},
		},
	}
}
