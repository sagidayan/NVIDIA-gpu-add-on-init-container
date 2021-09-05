package ops

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateRuntimeObject(runtimeClient runtimeclient.Client, cr runtimeclient.Object, logger logrus.FieldLogger) error {
	objName := fmt.Sprintf("%s %s", cr.GetObjectKind().GroupVersionKind(), cr.GetName())
	logger.Infof("Creating runtime object %s", objName)
	err := runtimeClient.Create(context.Background(), cr)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		logger.WithError(err).Errorf("Failed to create %s", objName)
		return err
	}

	if kerrors.IsAlreadyExists(err) {
		logger.Infof("Object %s already exists, skipping it", objName)
	}

	return nil
}

func GetCRasUnstructuredObjectFromAlmExample(almExample string, logger logrus.FieldLogger) (*unstructured.Unstructured, error) {
	var crJson unstructured.UnstructuredList
	err := json.Unmarshal([]byte(almExample), &crJson.Items)
	if err != nil {
		logger.WithError(err).Error("Failed to unmarshal alm example")
		return nil, err
	}

	if len(crJson.Items) < 1 {
		message := fmt.Sprintf("list crs is empty, something gone wrong. alm example %s", almExample)
		logger.Error(message)
		return nil, fmt.Errorf(message)
	}

	return &crJson.Items[0], nil
}
