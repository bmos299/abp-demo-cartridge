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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	knkafkabindings "knative.dev/eventing-contrib/kafka/source/pkg/apis/bindings/v1alpha1"
	knkafkasource "knative.dev/eventing-contrib/kafka/source/pkg/apis/sources/v1alpha1"
	knmessaging "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	knduckv1 "knative.dev/pkg/apis/duck/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	basev1beta1 "github.ibm.com/automation-base-pak/abp-base-operator/api/v1beta1"
)

func (r *IAFDemoReconciler) reconcileKnative(recctx *reconcileContext, deployedName string) error {
	log := r.Log.WithValues("iafdemo", recctx.req.NamespacedName)
	namespace := recctx.iafdemo.Namespace

	// Get the CartridgeRequirements to extract some information from its status
	cartridgeReqInstance := &basev1beta1.CartridgeRequirements{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: iafCartridgeReqInstanceName, Namespace: namespace}, cartridgeReqInstance)
	if err != nil && errors.IsNotFound(err) {
		return fmt.Errorf("Failed to find CartridgeRequirements %s in Namespace %s: %w", iafCartridgeReqInstanceName, namespace, err)
	} else if err != nil {
		return err
	}

	kafkaBootstrapServers := ""
	if cartridgeReqInstance.Status.Components != nil && cartridgeReqInstance.Status.Components.Kafka != nil {
		endpoints := cartridgeReqInstance.Status.Components.Kafka.Endpoints
		for _, endpoint := range endpoints {
			if endpoint.Name == "internal-service-plain" {
				kafkaBootstrapServers = endpoint.BootstrapServers
			}
		}
	} else {
		return fmt.Errorf("Failed to get KafkaBootstrap servers from CartridgeRequirements %s in Namespace %s", iafCartridgeReqInstanceName, namespace)
	}

	// Create KafkaSource if not present
	kafkaSource := &knkafkasource.KafkaSource{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: deployedName, Namespace: namespace}, kafkaSource)
	if err != nil && errors.IsNotFound(err) {
		log.Info("KafkaSource " + deployedName + " not found. Creating...")
		kafkaSource = newKafkaSource(deployedName, namespace, kafkaBootstrapServers)
		err = ctrl.SetControllerReference(recctx.iafdemo, kafkaSource, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}
		err = r.Create(*recctx.ctx, kafkaSource)
		if err != nil {
			return fmt.Errorf("Failed to create new KafkaSource %s in Namespace %s: %s", kafkaSource.Name, kafkaSource.Namespace, err)
		}
	} else if err != nil {
		return err
	}

	// Create InMemoryChannel if not present
	inMemoryChannel := &knmessaging.InMemoryChannel{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: deployedName, Namespace: namespace}, inMemoryChannel)
	if err != nil && errors.IsNotFound(err) {
		log.Info("InMemoryChannel " + deployedName + " not found. Creating...")
		inMemoryChannel = newInMemoryChannel(deployedName, namespace)
		err = ctrl.SetControllerReference(recctx.iafdemo, inMemoryChannel, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}
		err = r.Create(*recctx.ctx, inMemoryChannel)
		if err != nil {
			return fmt.Errorf("Failed to create new InMemoryChannel %s in Namespace %s: %s", inMemoryChannel.Name, inMemoryChannel.Namespace, err)
		}
	} else if err != nil {
		return err
	}

	// Create Subscription if not present
	subscription := &knmessaging.Subscription{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: deployedName, Namespace: namespace}, subscription)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Subscription " + deployedName + " not found. Creating...")
		subscription = newKnativeSubscription(deployedName, namespace)
		err = ctrl.SetControllerReference(recctx.iafdemo, subscription, r.Scheme)
		if err != nil {
			return fmt.Errorf("Failed to set controller reference: %s", err)
		}
		err = r.Create(*recctx.ctx, subscription)
		if err != nil {
			return fmt.Errorf("Failed to create new Subscription %s in Namespace %s: %s", subscription.Name, subscription.Namespace, err)
		}
	} else if err != nil {
		return err
	}
	return nil
}

func newKafkaSource(deployedName, namespace, kafkaBootstrapServers string) *knkafkasource.KafkaSource {
	return &knkafkasource.KafkaSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deployedName,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Spec: knkafkasource.KafkaSourceSpec{
			KafkaAuthSpec: knkafkabindings.KafkaAuthSpec{
				BootstrapServers: []string{kafkaBootstrapServers},
			},
			Topics:        []string{eventProcessorRiskTopic},
			ConsumerGroup: deployedName,
			Sink: &knduckv1.Destination{
				Ref: &knduckv1.KReference{
					Kind:       "InMemoryChannel",
					Namespace:  namespace,
					Name:       deployedName,
					APIVersion: "messaging.knative.dev/v1alpha1",
				},
			},
		},
	}
}

func newInMemoryChannel(deployedName, namespace string) *knmessaging.InMemoryChannel {
	return &knmessaging.InMemoryChannel{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deployedName,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Spec: knmessaging.InMemoryChannelSpec{},
	}
}

func newKnativeSubscription(deployedName, namespace string) *knmessaging.Subscription {
	return &knmessaging.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deployedName,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Spec: knmessaging.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				Kind:       "InMemoryChannel",
				Namespace:  namespace,
				Name:       deployedName,
				APIVersion: "messaging.knative.dev/v1alpha1",
			},
			Subscriber: &knduckv1.Destination{
				Ref: &knduckv1.KReference{
					Kind:       "Service",
					Namespace:  namespace,
					Name:       deployedName,
					APIVersion: "v1",
				},
			},
		},
	}
}
