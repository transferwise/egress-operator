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

package main

import (
	"flag"
	"fmt"
	"os"

	egressv1 "github.com/monzo/egress-operator/api/v1"
	"github.com/monzo/egress-operator/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const namespace = "egress-operator-system"

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = egressv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsPort int
	var metricsHost string
	var healthPort int
	var healthHost string
	var enableLeaderElection bool
	var leaderElectionNamespace string
	flag.IntVar(&metricsPort, "metrics-port", 8383, "The port the metric endpoint binds to.")
	flag.StringVar(&metricsHost, "metrics-host", "0.0.0.0", "The network interface the metric endpoint binds to.")
	flag.IntVar(&healthPort, "health-port", 8080, "The port the health endpoint binds to.")
	flag.StringVar(&healthHost, "health-host", "0.0.0.0", "The network interface to listen to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&leaderElectionNamespace, "leader-election-namespace", namespace, "Leader election namespace.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		HealthProbeBindAddress:  fmt.Sprintf("%s:%d", healthHost, healthPort),
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: leaderElectionNamespace,
		Namespace:               namespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ExternalServiceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ExternalService"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ExternalService")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "Error starting health check service")
		os.Exit(1)
	}

	if err = mgr.AddReadyzCheck("ready", healthz.Ping); err != nil {
		setupLog.Error(err, "Error starting readiness check service")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
