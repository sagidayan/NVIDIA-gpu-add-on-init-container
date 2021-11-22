package ops

import (
	"context"
	"fmt"
	"strings"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmv1client "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const almExamples = "alm-examples"

func GetCsvWithPrefix(olmClient *olmv1client.OperatorsV1alpha1Client, namespace string, prefix string) (*operatorsv1alpha1.ClusterServiceVersion, error) {
	csvs, err := olmClient.ClusterServiceVersions(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, csv := range csvs.Items {
		if strings.HasPrefix(csv.Name, prefix) {
			return &csv, nil
		}
	}
	return nil, fmt.Errorf("csv with prefix %s not found in %s", prefix, namespace)
}

func getAlmExamples(csv *operatorsv1alpha1.ClusterServiceVersion) (string, error) {
	annotations := csv.ObjectMeta.GetAnnotations()
	if val, ok := annotations[almExamples]; ok {
		return val, nil
	}

	return "", fmt.Errorf("%s not found in given csv %v", almExamples, csv)
}

func GetAlmExamples(olmClient *olmv1client.OperatorsV1alpha1Client, logger logrus.FieldLogger, namespace, prefix string) (string, error) {
	csv, err := GetCsvWithPrefix(olmClient, namespace, prefix)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get %s csv", prefix)
		return "", err
	}

	almExample, err := getAlmExamples(csv)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get %s alm example", prefix)
		return "", err
	}
	return almExample, nil
}
