module github.com/openshift/assisted-installer

go 1.16

require (
	github.com/NVIDIA/gpu-operator v1.8.1
	github.com/go-openapi/swag v0.19.5
	github.com/openshift/cluster-nfd-operator v0.0.0-20210901165408-adb87ce0d9b7
	github.com/sirupsen/logrus v1.8.1
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	sigs.k8s.io/controller-runtime v0.8.2
)

replace github.com/openshift/api => github.com/openshift/api v3.9.1-0.20191111211345-a27ff30ebf09+incompatible
