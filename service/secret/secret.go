package secret

import (
	"context"
	"encoding/base64"
	"sidecar/models"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func WatchSecrets(k8sClient *kubernetes.Clientset, namepspace string, opt v1.ListOptions, path string, getFiles map[string]models.FileInfo) error {
	secretClient := k8sClient.CoreV1().Secrets(namepspace)
	secrets, err := secretClient.List(context.TODO(), opt)
	if err != nil {
		return err
	}
	for _, item := range secrets.Items {
		resourceName := item.Name
		resourceVersion := item.ResourceVersion
		resourceUID := item.UID
		for name, contentBase64 := range item.Data {
			content, err := base64.StdEncoding.DecodeString(string(contentBase64))
			if err != nil {
				return err
			}
			file := models.FileInfo{
				ResourceName:    resourceName,
				ResourceUID:     resourceUID,
				ResourceVersion: resourceVersion,
				Content:         string(content),
			}
			getFiles[path+"/"+name] = file
		}
	}
	return nil
}
