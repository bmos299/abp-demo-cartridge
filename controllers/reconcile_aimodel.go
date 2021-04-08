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
	"os"
	"path/filepath"
	"strings"

	aiv1 "github.ibm.com/automation-base-pak/abp-ai-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	// aiv1 "github.ibm.com/automation-base-pak/abp-ai-operator/api/v1alpha1"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	aiModelInstanceName        = "anomaly-classifier"
	aiModelInstanceVersion     = "1.0.0"
	aiAPIVersion               = "1.0.0"
	aiModelInstanceDescription = "Sample AI Model that categorises risk"
	aiModelInstanceType        = "tensorflow"
	aiModelSourceSecret        = "github-secret"
	aiModelInstancePath        = "models/anomaly-classifier"

	aiModelInstanceBucket         = "iaf-ai"
	aiModelStorageSecret          = "minio-secret"
	aiModelKFSecret               = "kfserving-secret"
	aiKFServingRuntime            = "kfservingruntime"
	aiKFServingRuntimeType        = "serving"
	aiKFServingRuntimePlatform    = "kfserving"
	aiKFServingRuntimeDescription = "KFServing runtime to deploy models"

	aiDeploymentInstanceName = "anomaly-classifier"
)

func (r *IAFDemoReconciler) setupAIModels(recctx *reconcileContext) (bool, error) {
	r.Log.Info("Setting up AI Custom Resources")
	es := false
	namespace := recctx.iafdemo.Namespace
	authSecret := &corev1.Secret{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: aiModelStorageSecret, Namespace: namespace}, authSecret)
	if err != nil && errors.IsNotFound(err) {
		err = fmt.Errorf("Failed to find Secret %s in Namespace %s: %w", aiModelStorageSecret, namespace, err)
		return es, err
	} else if err != nil {
		return es, err
	}
	username, ok := authSecret.Data["AWS_ACCESS_KEY_ID"]
	if !ok {
		err = fmt.Errorf("Failed to get key 'AWS_ACCESS_KEY_ID' in Secret %s in Namespace %s: %w", aiModelStorageSecret, namespace, err)
		return es, err
	}
	accessKeyID := strings.TrimSuffix(string(username), "\n")
	password, ok := authSecret.Data["AWS_SECRET_ACCESS_KEY"]
	if !ok {
		err = fmt.Errorf("Failed to get key 'AWS_SECRET_ACCESS_KEY' in Secret %s in Namespace %s: %w", aiModelStorageSecret, namespace, err)
		return es, err
	}
	secretAccessKey := strings.TrimSuffix(string(password), "\n")
	url, ok := authSecret.Annotations["serving.kubeflow.org/s3-endpoint"]
	if !ok {
		err = fmt.Errorf("Failed to get endpoint annotation in Secret %s in Namespace %s: %w", aiModelStorageSecret, namespace, err)
		return es, err
	}
	endpoint := strings.TrimSuffix(string(url), "\n")

	useSSL := false
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		fmt.Println(err)
		return false, err
	}
	r.Log.Info("Copying bundled AIModels to Minio bucket")
	found, err := minioClient.BucketExists(context.Background(), aiModelInstanceBucket)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	if !found {
		// Create a bucket at region 'us-east-1' with object locking enabled.
		err = minioClient.MakeBucket(context.Background(), aiModelInstanceBucket, minio.MakeBucketOptions{})
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		err = filepath.Walk("models",
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					file, err := os.Open(path)
					if err != nil {
						fmt.Println(err)
						return err
					}
					defer file.Close()
					fileStat, err := file.Stat()
					if err != nil {
						fmt.Println(err)
						return err
					}
					_, err = minioClient.PutObject(context.Background(), aiModelInstanceBucket, path, file, fileStat.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
					if err != nil {
						fmt.Println(err)
						return err
					}
				}

				return nil
			})
		if err != nil {
			return false, err
		}

	}

	r.Log.Info("Deploying the Models on Kubeflow")
	return r.deployModelOnKubeflow(recctx)
}

func (r *IAFDemoReconciler) getPredictorURL(recctx *reconcileContext) (string, error) {
	r.Log.Info("Get the AIDeployment endpoint")
	aideployment := &aiv1.AIDeployment{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: aiDeploymentInstanceName, Namespace: recctx.iafdemo.Namespace}, aideployment)
	if err == nil {
		readyCondition := aideployment.Status.Conditions.GetCondition("Ready")
		if readyCondition != nil {
			if readyCondition.Status == corev1.ConditionTrue {
				if len(aideployment.Status.Endpoint) > 0 {
					return aideployment.Status.Endpoint, nil
				}
				return "", fmt.Errorf("Invalid Inference Service Endpoint")
			} else {
				return "", fmt.Errorf("Failed to create Inference Service: %s", readyCondition.Reason)
			}
		}
		return "", fmt.Errorf("Inference Service is not Ready")
	}
	return "", fmt.Errorf("Inference Service Not Found")
}

func (r *IAFDemoReconciler) deployModelOnKubeflow(recctx *reconcileContext) (bool, error) {
	//Check if AIDeployment exists
	r.Log.Info("Checking if AIDeployment exists")
	aideployment := &aiv1.AIDeployment{}
	err := r.Get(*recctx.ctx, types.NamespacedName{Name: aiDeploymentInstanceName, Namespace: recctx.iafdemo.Namespace}, aideployment)
	if err == nil {
		readyCondition := aideployment.Status.Conditions.GetCondition("Ready")

		if readyCondition != nil {
			if readyCondition.IsUnknown() || strings.EqualFold(string(readyCondition.Reason), "Pending") {
				r.Log.Info("UNKNOWN")
				return true, nil
			}
			if readyCondition.Status == corev1.ConditionTrue {
				if len(aideployment.Status.Endpoint) > 0 {
					r.Log.Info(aideployment.Status.Endpoint)
					return false, nil
				}
				return false, fmt.Errorf("Invalid Inference Service Endpoint")
			} else {
				return false, fmt.Errorf("Failed to create Inference Service: %s", readyCondition.Reason)
			}
		}
		return true, nil
	} else if !errors.IsNotFound(err) {
		return false, err
	}
	r.Log.Info("AIDeployment not found. Creating the custom resources")

	//Check if Kubeflow secret exists
	kfSecret := &corev1.Secret{}
	err = r.Get(*recctx.ctx, types.NamespacedName{Name: aiModelKFSecret, Namespace: recctx.iafdemo.Namespace}, kfSecret)
	if err != nil && errors.IsNotFound(err) {
		err = fmt.Errorf("Failed to find Secret %s in Namespace %s: %w", aiModelKFSecret, recctx.iafdemo.Namespace, err)
		return false, err
	} else if err != nil {
		return false, err
	}
	r.Log.Info("Creating AIRuntime")
	airuntime := &aiv1.AIRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      aiKFServingRuntime,
			Namespace: recctx.iafdemo.Namespace,
			Annotations: map[string]string{
				"com.ibm.automation.cartridge": iafCartridgeInstanceName,
			},
		},
		Spec: aiv1.AIRuntimeSpec{
			Version: aiAPIVersion,
			License: aiv1.License{
				Accept: bool(recctx.iafdemo.Spec.License.Accept),
			},
			Description: aiKFServingRuntimeDescription,
			Type:        aiKFServingRuntimeType,
			Platform:    aiKFServingRuntimePlatform,
			Credentials: aiv1.Credentials{SecretName: aiModelKFSecret},
		},
	}

	err = ctrl.SetControllerReference(recctx.iafdemo, airuntime, r.Scheme)
	if err != nil {
		return false, fmt.Errorf("Failed to set controller reference: %s", err)
	}

	err = r.Create(*recctx.ctx, airuntime)
	if err != nil && !errors.IsAlreadyExists(err) {
		return false, fmt.Errorf("Failed to create AIRuntime instance: %s", err)
	}

	r.Log.Info("Creating AIModel")
	aimodel := &aiv1.AIModel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      aiModelInstanceName,
			Namespace: recctx.iafdemo.Namespace,
			Annotations: map[string]string{
				"com.ibm.automation.cartridge": iafCartridgeInstanceName,
			},
		},
		Spec: aiv1.AIModelSpec{
			Version: aiAPIVersion,
			License: aiv1.License{
				Accept: bool(recctx.iafdemo.Spec.License.Accept),
			},
			Description: aiModelInstanceDescription,
			Type:        aiModelInstanceType,
			Source:      aiv1.Credentials{SecretName: aiModelSourceSecret},
			Store:       aiv1.Credentials{SecretName: aiModelStorageSecret},
		},
	}

	err = ctrl.SetControllerReference(recctx.iafdemo, aimodel, r.Scheme)
	if err != nil {
		return false, fmt.Errorf("Failed to set controller reference: %s", err)
	}

	err = r.Create(*recctx.ctx, aimodel)
	if err != nil && !errors.IsAlreadyExists(err) {
		return false, fmt.Errorf("Failed to create AIModel instance: %s", err)
	}

	r.Log.Info("Creating AIDeployment")
	aideployment = &aiv1.AIDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      aiDeploymentInstanceName,
			Namespace: recctx.iafdemo.Namespace,
			Annotations: map[string]string{
				"com.ibm.automation.cartridge": iafCartridgeInstanceName,
			},
		},
		Spec: aiv1.AIDeploymentSpec{
			Version: aiAPIVersion,
			License: aiv1.License{
				Accept: bool(recctx.iafdemo.Spec.License.Accept),
			},
			Runtime: aiKFServingRuntime,
			Model:   aiv1.Model{Name: aiModelInstanceName, Version: aiModelInstanceVersion},
		},
	}

	err = ctrl.SetControllerReference(recctx.iafdemo, aideployment, r.Scheme)
	if err != nil {
		return false, fmt.Errorf("Failed to set controller reference: %s", err)
	}

	err = r.Create(*recctx.ctx, aideployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		return false, fmt.Errorf("Failed to create AIDeployment instance: %s", err)
	}

	return true, nil
}
