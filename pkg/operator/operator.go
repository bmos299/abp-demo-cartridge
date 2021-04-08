// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// Licensed Materials - Property of IBM
// 5900-AEO
//
// Copyright IBM Corp. 2020, 2021. All Rights Reserved.
//
// US Government Users Restricted Rights - Use, duplication, or
// disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
// ------------------------------------------------------ {COPYRIGHT-END} ---
package operator

import (
	"flag"
	"os"

	routev1 "github.com/openshift/api/route/v1"
	aiv1alpha1 "github.ibm.com/automation-base-pak/abp-ai-operator/api/v1alpha1"
	basev1beta1 "github.ibm.com/automation-base-pak/abp-base-operator/api/v1beta1"
	kafkav1beta1 "github.ibm.com/automation-base-pak/abp-base-operator/api/v1beta1kafka"
	corev1beta1 "github.ibm.com/automation-base-pak/abp-core-operator/api/v1beta1"
	democartridgev1 "github.ibm.com/automation-base-pak/abp-demo-cartridge/api/v1"
	"github.ibm.com/automation-base-pak/abp-demo-cartridge/controllers"
	"github.ibm.com/automation-base-pak/abp-demo-cartridge/pkg/config"
	epv1alpha1 "github.ibm.com/automation-base-pak/abp-eventprocessing/api/v1alpha1"
	epv1beta1 "github.ibm.com/automation-base-pak/abp-eventprocessing/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	knkafkasource "knative.dev/eventing-contrib/kafka/source/pkg/apis/sources/v1alpha1"
	knmessaging "knative.dev/eventing/pkg/apis/messaging/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// Start Operator resource controllers
func Start(cfg *config.Config) {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(democartridgev1.AddToScheme(scheme))
	utilruntime.Must(corev1beta1.AddToScheme(scheme))
	utilruntime.Must(basev1beta1.AddToScheme(scheme))
	utilruntime.Must(kafkav1beta1.AddToScheme(scheme))
	utilruntime.Must(aiv1alpha1.AddToScheme(scheme))
	utilruntime.Must(epv1alpha1.AddToScheme(scheme))
	utilruntime.Must(epv1beta1.AddToScheme(scheme))
	utilruntime.Must(routev1.AddToScheme(scheme))
	utilruntime.Must(knkafkasource.AddToScheme(scheme))
	utilruntime.Must(knmessaging.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme

	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "028d8a12.ibm.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.IAFDemoReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IAFDemo"),
		Scheme: mgr.GetScheme(),
		Cfg:    cfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IAFDemo")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
