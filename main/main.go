package main

import (
	"github.com/gpu-init-container/src/config"
	"github.com/gpu-init-container/src/ops"
	"github.com/sirupsen/logrus"

	gpuv1 "github.com/NVIDIA/gpu-operator/api/v1"
	nfdv1 "github.com/openshift/cluster-nfd-operator/api/v1"
	olmv1client "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	runtimeconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {

	config.ProcessArgs()

	logger := logrus.New()
	logger.Infof("Start running init container")

	logger.Info("Initializing clients")
	olmClient, runtimeClient, err := initClients()
	if err != nil {
		logger.WithError(err).Fatal("Failed initializing clients")
	}

	err = createNfdCr(olmClient, runtimeClient, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed creating NFD CR")
	}

	err = createClusterPolicyCr(olmClient, runtimeClient, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed creating ClusterPolicy CR")
	}
}

func initClients() (*olmv1client.OperatorsV1alpha1Client, runtimeclient.Client, error) {
	scheme := runtime.NewScheme()
	var runtimeClient runtimeclient.Client
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		return nil, nil, err
	}

	err = nfdv1.AddToScheme(scheme)
	if err != nil {
		return nil, nil, err
	}
	err = gpuv1.AddToScheme(scheme)
	if err != nil {
		return nil, nil, err
	}

	runtimeClient, err = runtimeclient.New(runtimeconfig.GetConfigOrDie(), runtimeclient.Options{Scheme: scheme})
	if err != nil {
		return nil, nil, err
	}

	olmClient, err := olmv1client.NewForConfig(runtimeconfig.GetConfigOrDie())
	if err != nil {
		return nil, nil, err
	}

	return olmClient, runtimeClient, nil
}

func createNfdCr(olmClient *olmv1client.OperatorsV1alpha1Client, runtimeClient runtimeclient.Client, logger logrus.FieldLogger) error {
	logger.Info("Creating nfd cr")
	nfdExample, err := ops.GetAlmExamples(olmClient, logger, config.GlobalConfig.Namespace, config.GlobalConfig.NfdPrefix)
	if err != nil {
		logger.WithError(err).Error("Failed to get nfd alm example")
		return err
	}

	nfdCr, err := ops.GetCRasUnstructuredObjectFromAlmExample(nfdExample, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get nfd cr from alm examples")
		return err
	}

	nfdCr.SetNamespace(config.GlobalConfig.Namespace)
	err = ops.CreateRuntimeObject(runtimeClient, nfdCr, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to push nfd")
		return err
	}

	logger.Info("Done creating nfd cr")
	return nil
}


func createClusterPolicyCr(olmClient *olmv1client.OperatorsV1alpha1Client, runtimeClient runtimeclient.Client, logger logrus.FieldLogger) error {
	logger.Info("Creating clusterPolicy cr")
	clusterPolicyExample, err := ops.GetAlmExamples(olmClient, logger, config.GlobalConfig.Namespace, config.GlobalConfig.GpuPrefix)
	if err != nil {
		logger.WithError(err).Error("Failed to get clusterPolicy alm example")
		return err
	}

	clusterPolicy, err := ops.GetCRasUnstructuredObjectFromAlmExample(clusterPolicyExample, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to get clusterPolicy cr from alm examples")
		return err
	}
	clusterPolicy.SetNamespace(config.GlobalConfig.Namespace)
	err = ops.CreateRuntimeObject(runtimeClient, clusterPolicy, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to push clusterPolicy")
		return err
	}

	logger.Info("Done creating clusterPolicy cr")
	return nil
}
