package configmap

import (
	"context"
	"fmt"
	"sidecar/models"

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

var (
	Logger      *zap.Logger
	nilFileInfo = models.FileInfo{}
)

func ListConfigMaps(k8sClient *kubernetes.Clientset, resource models.Resource, getFiles map[string]models.FileInfo) error {
	configMapClient := k8sClient.CoreV1().ConfigMaps(resource.Namespace)
	labelset := make(map[string]string)
	for _, label := range resource.Labels {
		labelset[label.Name] = label.Value
	}
	listoption := v1.ListOptions{LabelSelector: labels.Set(labelset).String()}
	configMaps, err := configMapClient.List(context.TODO(), listoption)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "WatchConfigMaps"))
		return err
	}

	for _, item := range configMaps.Items {
		resourceName := item.Name
		resourceVersion := item.ResourceVersion
		resourceUID := item.UID
		for name, content := range item.Data {
			fullName := resource.Path + "/" + name
			if getFiles[fullName].ResourceName != "" {
				Logger.Warn(fmt.Sprintf("Ingnore file %s in namesapce %s, configmap %s, resouce UID %s and resource version %s because file %s is already seen.",
					resource.Namespace, name, resourceName, resourceUID, resourceVersion, fullName), zap.String("method", "WatchConfigMaps"))
				continue
			}
			file := models.FileInfo{
				ResourceName:    resourceName,
				ResourceUID:     resourceUID,
				ResourceVersion: resourceVersion,
				Content:         []byte(content),
				Index:           resource.Index,
			}
			getFiles[fullName] = file
		}
	}
	return nil
}
