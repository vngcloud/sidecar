package configmap

import (
	"context"
	"sidecar/models"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func WatchConfigMaps(k8sClient *kubernetes.Clientset, namepspace string, opt v1.ListOptions, path string, getFiles map[string]models.FileInfo) error {
	configMapClient := k8sClient.CoreV1().ConfigMaps(namepspace)
	configMaps, err := configMapClient.List(context.TODO(), opt)
	if err != nil {
		return err
	}
	for _, item := range configMaps.Items {
		resourceName := item.Name
		resourceVersion := item.ResourceVersion
		resourceUID := item.UID
		for name, content := range item.Data {
			file := models.FileInfo{
				ResourceName:    resourceName,
				ResourceUID:     resourceUID,
				ResourceVersion: resourceVersion,
				Content:         content,
			}
			getFiles[path+"/"+name] = file
		}
	}
	return nil
}
