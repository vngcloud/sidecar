package secret

import (
	"context"
	"encoding/base64"
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

func WatchSecrets(k8sClient *kubernetes.Clientset, namepspace string, opt v1.ListOptions, path string, getFiles map[string]models.FileInfo) error {
	secretClient := k8sClient.CoreV1().Secrets(namepspace)
	secrets, err := secretClient.List(context.TODO(), opt)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "WatchSecrets"))
		return err
	}
	for _, item := range secrets.Items {
		resourceName := item.Name
		resourceVersion := item.ResourceVersion
		resourceUID := item.UID
		for name, contentBase64 := range item.Data {
			fullName := path + "/" + name
			if getFiles[fullName] != nilFileInfo {
				Logger.Warn(fmt.Sprintf("Ingnore file %s in secret %s, resouce UID %s and resource version %s because file %s is exits.",
					name, resourceName, resourceUID, resourceVersion, fullName), zap.String("method", "WatchSecrets"))
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
				Content:         string(content),
			}
			getFiles[fullName] = file
		}
	}
	return nil
}
