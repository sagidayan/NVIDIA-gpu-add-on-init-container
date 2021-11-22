module github.com/rh-ecosystem-edge/NVIDIA-gpu-add-on-init-container

go 1.16

require (
	github.com/NVIDIA/gpu-operator v1.8.1
	github.com/openshift/cluster-nfd-operator v0.0.0-20210901165408-adb87ce0d9b7
	github.com/operator-framework/api v0.9.2
	github.com/operator-framework/operator-lifecycle-manager v0.18.3
	github.com/sirupsen/logrus v1.8.1
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/openshift/api => github.com/openshift/api v3.9.1-0.20191111211345-a27ff30ebf09+incompatible
