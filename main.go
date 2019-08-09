package main

import (
	"flag"
	"net/http"

	"github.com/golang/glog"
	"github.com/nirmata/kyverno/pkg/annotations"
	v1alpha1 "github.com/nirmata/kyverno/pkg/client/listers/policy/v1alpha1"
	"github.com/nirmata/kyverno/pkg/config"
	controller "github.com/nirmata/kyverno/pkg/controller"
	client "github.com/nirmata/kyverno/pkg/dclient"
	event "github.com/nirmata/kyverno/pkg/event"
	gencontroller "github.com/nirmata/kyverno/pkg/gencontroller"
	kyvernoprometheus "github.com/nirmata/kyverno/pkg/prometheus"
	"github.com/nirmata/kyverno/pkg/sharedinformer"
	"github.com/nirmata/kyverno/pkg/utils"
	"github.com/nirmata/kyverno/pkg/violation"
	"github.com/nirmata/kyverno/pkg/webhooks"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/sample-controller/pkg/signals"
)

var (
	kubeconfig        string
	serverIP          string
	filterK8Resources string
	cpu               bool
	memory            bool
	webhookTimeout    int
)

func main() {
	defer glog.Flush()
	printVersionInfo()
	prof = enableProfiling(cpu, memory)

	clientConfig, err := createClientConfig(kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v\n", err)
	}

	client, err := client.NewClient(clientConfig)
	if err != nil {
		glog.Fatalf("Error creating client: %v\n", err)
	}

	policyInformerFactory, err := sharedinformer.NewSharedInformerFactory(clientConfig)
	if err != nil {
		glog.Fatalf("Error creating policy sharedinformer: %v\n", err)
	}

	webhookRegistrationClient, err := webhooks.NewWebhookRegistrationClient(clientConfig, client, serverIP, int32(webhookTimeout))
	if err != nil {
		glog.Fatalf("Unable to register admission webhooks on cluster: %v\n", err)
	}

	if err = webhookRegistrationClient.Register(); err != nil {
		glog.Fatalf("Failed registering Admission Webhooks: %v\n", err)
	}

	kubeInformer := utils.NewKubeInformerFactory(clientConfig)
	eventController := event.NewEventController(client, policyInformerFactory)
	violationBuilder := violation.NewPolicyViolationBuilder(client, policyInformerFactory, eventController)
	annotationsController := annotations.NewAnnotationControler(client)
	policyController := controller.NewPolicyController(
		client,
		policyInformerFactory,
		violationBuilder,
		eventController,
		filterK8Resources)

	genControler := gencontroller.NewGenController(client, eventController, policyInformerFactory, violationBuilder, kubeInformer.Core().V1().Namespaces(), annotationsController)
	tlsPair, err := initTLSPemPair(clientConfig, client)
	if err != nil {
		glog.Fatalf("Failed to initialize TLS key/certificate pair: %v\n", err)
	}

	server, err := webhooks.NewWebhookServer(client, tlsPair, policyInformerFactory, eventController, violationBuilder, annotationsController, webhookRegistrationClient, filterK8Resources)
	if err != nil {
		glog.Fatalf("Unable to create webhook server: %v\n", err)
	}

	stopCh := signals.SetupSignalHandler()

	policyInformerFactory.Run(stopCh)
	kubeInformer.Start(stopCh)
	eventController.Run(stopCh)
	genControler.Run(stopCh)
	annotationsController.Run(stopCh)
	if err = policyController.Run(stopCh); err != nil {
		glog.Fatalf("Error running PolicyController: %v\n", err)
	}

	server.RunAsync()

	go startPrometheus(policyInformerFactory.GetLister())

	<-stopCh
	genControler.Stop()
	eventController.Stop()
	annotationsController.Stop()
	policyController.Stop()
	disableProfiling(prof)
	server.Stop()
}

func init() {
	// profiling feature gate
	// cpu and memory profiling cannot be enabled at same time
	// if both cpu and memory are enabled
	// by default is to profile cpu
	flag.BoolVar(&cpu, "cpu", false, "cpu profilling feature gate, default to false || cpu and memory profiling cannot be enabled at the same time")
	flag.BoolVar(&memory, "memory", false, "memory profilling feature gate, default to false || cpu and memory profiling cannot be enabled at the same time")

	flag.IntVar(&webhookTimeout, "webhooktimeout", 2, "timeout for webhook configurations")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&serverIP, "serverIP", "", "IP address where Kyverno controller runs. Only required if out-of-cluster.")
	flag.StringVar(&filterK8Resources, "filterK8Resources", "", "k8 resource in format [kind,namespace,name] where policy is not evaluated by the admission webhook. example --filterKind \"[Deployment, kyverno, kyverno]\" --filterKind \"[Deployment, kyverno, kyverno],[Events, *, *]\"")
	config.LogDefaultFlags()
	flag.Parse()
}

func startPrometheus(policyLister v1alpha1.PolicyLister) {
	pm := kyvernoprometheus.InitPrometheusMetrics("kyverno")

	metricsManager := kyvernoprometheus.NewMetricsManager(policyLister)
	pm.AuditPolicy.Set(metricsManager.CountAuditPolicy())

	http.Handle("/metrics", promhttp.Handler())
	glog.Fatal(http.ListenAndServe(":8080", nil))
}
