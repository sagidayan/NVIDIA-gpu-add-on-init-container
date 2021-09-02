package main

import (
	"context"
	"github.com/go-openapi/swag"
	"github.com/sirupsen/logrus"
	"log"
	runtimeconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	gpuv1 "github.com/NVIDIA/gpu-operator/api/v1"
	nfdv1 "github.com/openshift/cluster-nfd-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)




func main() {
	logger := logrus.New()

	logger.Infof("Start running init")


	scheme := runtime.NewScheme()
	var runtimeClient runtimeclient.Client
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		log.Fatal(err)
	}

	err = nfdv1.AddToScheme(scheme)
	if err != nil {
		log.Fatal(err)
	}
	err = gpuv1.AddToScheme(scheme)
	if err != nil {
		log.Fatal(err)
	}

	runtimeClient, err = runtimeclient.New(runtimeconfig.GetConfigOrDie(), runtimeclient.Options{Scheme: scheme})
	if err != nil {
		log.Fatal(err)
	}


	logger.Info("Creating nfd cr")
	nfdCr := &nfdv1.NodeFeatureDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nfd-instance",
			Namespace: "redhat-gpu-operator",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "NodeFeatureDiscovery",
			APIVersion: "nfd.openshift.io/v1",
		},
		Spec: nfdv1.NodeFeatureDiscoverySpec{ Operand: nfdv1.OperandSpec{
			Namespace: "openshift-nfd",
			Image:           "registry.redhat.io/openshift4/ose-node-feature-discovery@sha256:a3ed882e2d6e227d1d746fcefa8e129fec8bd1843d8dbece9888986474af7da6",
			ImagePullPolicy: "Always" },
			Instance: "",
			WorkerConfig: &nfdv1.ConfigMap{
				ConfigData: "core:\n#  labelWhiteList:\n#  noPublish: false\n  sleepInterval: 60s\n#  sources: [all]\n#  klog:\n#    addDirHeader: false\n#    alsologtostderr: false\n#    logBacktraceAt:\n#    logtostderr: true\n#    skipHeaders: false\n#    stderrthreshold: 2\n#    v: 0\n#    vmodule:\n##   NOTE: the following options are not dynamically run-time configurable\n##         and require a nfd-worker restart to take effect after being changed\n#    logDir:\n#    logFile:\n#    logFileMaxSize: 1800\n#    skipLogHeaders: false\nsources:\n#  cpu:\n#    cpuid:\n##     NOTE: whitelist has priority over blacklist\n#      attributeBlacklist:\n#        - \"BMI1\"\n#        - \"BMI2\"\n#        - \"CLMUL\"\n#        - \"CMOV\"\n#        - \"CX16\"\n#        - \"ERMS\"\n#        - \"F16C\"\n#        - \"HTT\"\n#        - \"LZCNT\"\n#        - \"MMX\"\n#        - \"MMXEXT\"\n#        - \"NX\"\n#        - \"POPCNT\"\n#        - \"RDRAND\"\n#        - \"RDSEED\"\n#        - \"RDTSCP\"\n#        - \"SGX\"\n#        - \"SSE\"\n#        - \"SSE2\"\n#        - \"SSE3\"\n#        - \"SSE4.1\"\n#        - \"SSE4.2\"\n#        - \"SSSE3\"\n#      attributeWhitelist:\n#  kernel:\n#    kconfigFile: \"/path/to/kconfig\"\n#    configOpts:\n#      - \"NO_HZ\"\n#      - \"X86\"\n#      - \"DMI\"\n  pci:\n    deviceClassWhitelist:\n      - \"0200\"\n      - \"03\"\n      - \"12\"\n    deviceLabelFields:\n#      - \"class\"\n      - \"vendor\"\n#      - \"device\"\n#      - \"subsystem_vendor\"\n#      - \"subsystem_device\"\n#  usb:\n#    deviceClassWhitelist:\n#      - \"0e\"\n#      - \"ef\"\n#      - \"fe\"\n#      - \"ff\"\n#    deviceLabelFields:\n#      - \"class\"\n#      - \"vendor\"\n#      - \"device\"\n#  custom:\n#    - name: \"my.kernel.feature\"\n#      matchOn:\n#        - loadedKMod: [\"example_kmod1\", \"example_kmod2\"]\n#    - name: \"my.pci.feature\"\n#      matchOn:\n#        - pciId:\n#            class: [\"0200\"]\n#            vendor: [\"15b3\"]\n#            device: [\"1014\", \"1017\"]\n#        - pciId :\n#            vendor: [\"8086\"]\n#            device: [\"1000\", \"1100\"]\n#    - name: \"my.usb.feature\"\n#      matchOn:\n#        - usbId:\n#          class: [\"ff\"]\n#          vendor: [\"03e7\"]\n#          device: [\"2485\"]\n#        - usbId:\n#          class: [\"fe\"]\n#          vendor: [\"1a6e\"]\n#          device: [\"089a\"]\n#    - name: \"my.combined.feature\"\n#      matchOn:\n#        - pciId:\n#            vendor: [\"15b3\"]\n#            device: [\"1014\", \"1017\"]\n#          loadedKMod : [\"vendor_kmod1\", \"vendor_kmod2\"]\n",
			},
			CustomConfig: nfdv1.ConfigMap{
				ConfigData: "#    - name: \"more.kernel.features\"\n#      matchOn:\n#      - loadedKMod: [\"example_kmod3\"]\n#    - name: \"more.features.by.nodename\"\n#      value: customValue\n#      matchOn:\n#      - nodename: [\"special-.*-node-.*\"]\n",
			},
		},
	}

	logger.Info("Pushing nfd cr to cluster")
	err = runtimeClient.Create(context.Background(), nfdCr)
	if err != nil {
		logger.WithError(err).Error("Failed to create nfd cr")
		log.Fatal(err)
	}

	err = createClusterPolicy(runtimeClient, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to create cluster policy cr")
		log.Fatal(err)
	}
}


func createClusterPolicy(runtimeClient runtimeclient.Client, logger logrus.FieldLogger) error {
	logger.Info("Creating cluster policy cr")
	clusterPolicyCr := &gpuv1.ClusterPolicy{
		ObjectMeta: metav1.ObjectMeta{
		Name:      "gpu-cluster-policy",
	},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterPolicy",
			APIVersion: "nvidia.com/v1",
		},
		Spec: gpuv1.ClusterPolicySpec{
			Operator:            gpuv1.OperatorSpec{
				DefaultRuntime: "crio", 
				InitContainer: gpuv1.InitContainerSpec{
				Repository:       "nvcr.io/nvidia",
				Image:            "cuda",
				Version:          "sha256:15674e5c45c97994bc92387bad03a0d52d7c1e983709c471c4fecc8e806dbdce",
			}},
			Daemonsets:          gpuv1.DaemonsetsSpec{ Tolerations: []corev1.Toleration{corev1.Toleration{
				Key:               "nvidia.com/gpu",
				Operator:          "Exists",
				Effect:            "NoSchedule",
			}},
			PriorityClassName: "system-node-critical",
			},
			Driver:              gpuv1.DriverSpec{
				Enabled:          swag.Bool(true),
				GPUDirectRDMA:    &gpuv1.GPUDirectRDMASpec{Enabled: swag.Bool(true)},
				Repository:       "nvcr.io/nvidia",
				Image:            "driver",
				Version:          "sha256:a62de5e843a41c65cf837e7db5f5b675d03fa2de05e981a859b114336cf183e3",
				ImagePullSecrets: []string{},
				Manager:          gpuv1.DriverManagerSpec{
					Repository:       "nvcr.io/nvidia/cloud-native",
					Image:            "k8s-driver-manager",
					Version:          "sha256:907ab0fc008bb90149ed059ac3a8ed3d19ae010d52c58c0ddbafce45df468d5b",
					ImagePullSecrets: []string{},
					Env:              []corev1.EnvVar{
						{
							Name:  "DRAIN_USE_FORCE",
							Value: "false",
						},
						{
							Name: "DRAIN_POD_SELECTOR_LABEL",
							Value: "",
						},
						{
							Name: "DRAIN_TIMEOUT_SECONDS",
							Value: "0s",
						},
						{
							Name: "DRAIN_DELETE_EMPTYDIR_DATA",
							Value: "false",
						},
					},
				},
			},
			Toolkit:             gpuv1.ToolkitSpec{
				Enabled:          swag.Bool(true),
				Repository:       "nvcr.io/nvidia/k8s",
				Image:            "container-toolkit",
				Version:          "sha256:8f9517b4c83b8730c40134df385088be41519b585176c66727ff6f181ae5e703",
				ImagePullSecrets: []string{},
			},
			DevicePlugin:        gpuv1.DevicePluginSpec{
				Repository:       "nvcr.io/nvidia",
				Image:            "k8s-device-plugin",
				Version:          "sha256:85def0197f388e5e336b1ab0dbec350816c40108a58af946baa1315f4c96ee05",
				ImagePullSecrets: []string{},
				Env:              []corev1.EnvVar{
					{
						Name: "PASS_DEVICE_SPECS",
						Value: "true",
					},
					{
						Name: "FAIL_ON_INIT_ERROR",
						Value: "true",
					},
					{
						Name: "DEVICE_LIST_STRATEGY",
						Value: "envvar",
					},
					{
						Name: "DEVICE_ID_STRATEGY",
						Value: "uuid",
					},
					{
						Name: "NVIDIA_VISIBLE_DEVICES",
						Value: "all",
					},
					{
						Name: "NVIDIA_DRIVER_CAPABILITIES",
						Value: "all",
					},
				},
			},
			DCGMExporter:        gpuv1.DCGMExporterSpec{
				Repository:       "nvcr.io/nvidia/k8s",
				Image:            "dcgm-exporter",
				Version:          "sha256:e37404194fa2bc2275827411049422b93d1493991fb925957f170b4b842846ff",
				Env:              []corev1.EnvVar{
					{
						Name: "DCGM_EXPORTER_LISTEN",
						Value: ":9400",
					},
					{
						Name: "DCGM_EXPORTER_KUBERNETES",
						Value: "true",
					},
					{
						Name: "DCGM_EXPORTER_COLLECTORS",
						Value: "/etc/dcgm-exporter/dcp-metrics-included.csv",
					},
				},
			},
			DCGM:                gpuv1.DCGMSpec{
				Enabled:          swag.Bool(true),
				Repository:       "nvcr.io/nvidia/cloud-native",
				Image:            "dcgm",
				Version:          "sha256:28f334d6d5ca6e5cad2cf05a255989834128c952e3c181e6861bd033476d4b2c",
				HostPort:         5555,
			},
			NodeStatusExporter:  gpuv1.NodeStatusExporterSpec{
				Enabled:          swag.Bool(true),
				Repository:       "nvcr.io/nvidia/cloud-native",
				Image:            "gpu-operator-validator",
				Version:          "sha256:1cce434a1722288bacab5eaa5c194ca2bdbad55679ba871a2814556853339585",
			},
			GPUFeatureDiscovery: gpuv1.GPUFeatureDiscoverySpec{
				Repository:       "nvcr.io/nvidia",
				Image:            "gpu-feature-discovery",
				Version:          "sha256:bfc39d23568458dfd50c0c5323b6d42bdcd038c420fb2a2becd513a3ed3be27f",
				Env:              []corev1.EnvVar{
					{
						Name: "GFD_SLEEP_INTERVAL",
						Value: "60s",
					},
					{
						Name: "FAIL_ON_INIT_ERROR",
						Value: "true",
					},

				},
			},
			MIG:                 gpuv1.MIGSpec{Strategy: "single"},
			MIGManager:          gpuv1.MIGManagerSpec{
				Enabled:          swag.Bool(true),
				Repository:       "nvcr.io/nvidia/cloud-native",
				Image:            "k8s-mig-manager",
				Version:          "sha256:77b8e58a54c222bee3cc56b2305d4cebfa60722c122858f94301e611f87d7fec",
				Env:              []corev1.EnvVar{
					{
						Name: "WITH_REBOOT",
						Value: "false",
					},
				},
			},
			Validator:           gpuv1.ValidatorSpec{
				Repository:       "nvcr.io/nvidia/cloud-native",
				Image:            "gpu-operator-validator",
				Version:          "sha256:1cce434a1722288bacab5eaa5c194ca2bdbad55679ba871a2814556853339585",
				Env:              []corev1.EnvVar{
					{
						Name: "WITH_WORKLOAD",
						Value: "true",
					},
			},
			},
		},
	}
	logger.Info("Pushing nfd cr to cluster")
	return runtimeClient.Create(context.Background(), clusterPolicyCr)
}