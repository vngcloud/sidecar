package configmap

import (
	"context"
	"fmt"
	"sidecar/models"

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	Logger      *zap.Logger
	nilFileInfo = models.FileInfo{}
)

func WatchConfigMaps(k8sClient *kubernetes.Clientset, namepspace string, opt v1.ListOptions, path string, getFiles map[string]models.FileInfo) error {
	configMapClient := k8sClient.CoreV1().ConfigMaps(namepspace)
	configMaps, err := configMapClient.List(context.TODO(), opt)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "WatchConfigMaps"))
		return err
	}
	for _, item := range configMaps.Items {
		resourceName := item.Name
		resourceVersion := item.ResourceVersion
		resourceUID := item.UID
		for name, content := range item.Data {
			fullName := path + "/" + name
			if getFiles[fullName] != nilFileInfo {
				Logger.Warn(fmt.Sprintf("Ingnore file %s in configmap %s, resouce UID %s and resource version %s because file %s is exits.",
					name, resourceName, resourceUID, resourceVersion, fullName), zap.String("method", "WatchConfigMaps"))
				continue
			}
			file := models.FileInfo{
				ResourceName:    resourceName,
				ResourceUID:     resourceUID,
				ResourceVersion: resourceVersion,
				Content:         content,
			}
			getFiles[fullName] = file
		}
	}
	return nil
}
