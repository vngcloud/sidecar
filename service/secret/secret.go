package secret

import (
	"context"
	"encoding/base64"
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

func ListSecrets(k8sClient *kubernetes.Clientset, resource models.Resource, getFiles map[string]models.FileInfo) error {
	secretClient := k8sClient.CoreV1().Secrets(resource.Namespace)
	labelset := make(map[string]string)
	for _, label := range resource.Labels {
		labelset[label.Name] = label.Value
	}
	listoption := v1.ListOptions{LabelSelector: labels.Set(labelset).String()}
	secrets, err := secretClient.List(context.TODO(), listoption)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "WatchSecrets"))
		return err
	}
	for _, item := range secrets.Items {
		resourceName := item.Name
		resourceVersion := item.ResourceVersion
		resourceUID := item.UID
		for name, contentBase64 := range item.Data {
			fullName := resource.Path + "/" + name
			if getFiles[fullName].ResourceName != "" {
				Logger.Warn(fmt.Sprintf("Ingnore file %s in namespace %s,secret %s, resouce UID %s and resource version %s because file %s is is already seen.",
					resource.Namespace, name, resourceName, resourceUID, resourceVersion, fullName), zap.String("method", "WatchSecrets"))
				continue
			}
			content, err := base64.StdEncoding.DecodeString(string(contentBase64))
			if err != nil {
				Logger.Error(err.Error(), zap.String("method", "WatchSecrets"))
				return err
			}
			file := models.FileInfo{
				ResourceName:    resourceName,
				ResourceUID:     resourceUID,
				ResourceVersion: resourceVersion,
				Content:         content,
				Namespace:       resource.Namespace,
			}
			getFiles[fullName] = file
		}
	}
	return nil
}
