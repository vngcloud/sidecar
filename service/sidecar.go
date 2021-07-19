package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"sidecar/models"
	"sidecar/service/configmap"
	"sidecar/service/secret"
	"time"

	validator "github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	nilFileInfo = models.FileInfo{}
	Logger      *zap.Logger
)

func Init(namespace string, sleepTime int, fileK8sConfig string, fileConfig string) error {
	yamlFile, err := ioutil.ReadFile(fileConfig)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "Init"))
		return err
	}
	var resources models.Resources
	err = yaml.Unmarshal(yamlFile, &resources)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "Init"))
		return err
	}
	validate := validator.New()
	//validate.RegisterStructValidation(StructLevelValidation, models.Resources{})
	err = validate.Struct(&resources)
	var configK8s *rest.Config
	if fileK8sConfig == "" {
		configK8s, err = rest.InClusterConfig()
	} else {
		configK8s, err = clientcmd.BuildConfigFromFlags("", fileK8sConfig)
	}
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "Init"))
		return err
	}
	Logger.Info(fmt.Sprintf("Config for cluster api at '%s' loaded.\n", configK8s.Host), zap.String("method", "Init"))
	clientK8s, err := kubernetes.NewForConfig(configK8s)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "Init"))
		return err
	}
	for _, resource := range resources.Resources {
		err := os.MkdirAll(resource.Path, 0777)
		if err != nil {
			Logger.Warn(err.Error(), zap.String("method", "Init"))
		}
	}
	configmap.Logger = Logger.With(zap.String("package", "configmap"))
	secret.Logger = Logger.With(zap.String("package", "secret"))
	return WatchResource(resources, clientK8s, namespace, sleepTime)
}

func WatchResource(resoures models.Resources, k8sClient *kubernetes.Clientset, namespace string, sleepTime int) error {
	presentFiles := make(map[string]models.FileInfo)
	for {
		getFiles := make(map[string]models.FileInfo)
		for _, resource := range resoures.Resources {
			labelset := make(map[string]string)
			for _, label := range resource.Labels {
				labelset[label.Name] = label.Value
			}
			opt := v1.ListOptions{LabelSelector: labels.Set(labelset).String()}
			if resource.Type == "both" || resource.Type == "configmap" {
				err := configmap.WatchConfigMaps(k8sClient, namespace, opt, resource.Path, getFiles)
				if err != nil {
					Logger.Error(err.Error(), zap.String("method", "WatchResource"))
					return err
				}
			}
			if resource.Type == "both" || resource.Type == "secret" {
				err := secret.WatchSecrets(k8sClient, namespace, opt, resource.Path, getFiles)
				if err != nil {
					Logger.Error(err.Error(), zap.String("method", "WatchResource"))
					return err
				}
			}
		}
		err := DiffFiles(presentFiles, getFiles)
		if err != nil {
			Logger.Error(err.Error(), zap.String("method", "WatchResource"))
			return err
		}
		presentFiles = getFiles
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return nil
}

func DiffFiles(oldFiles map[string]models.FileInfo, newFiles map[string]models.FileInfo) error {
	for fileName, fileInfo := range newFiles {
		if oldFiles[fileName] == nilFileInfo {
			f, err := os.Create(fileName)
			if err != nil {
				Logger.Error(err.Error(), zap.String("method", "DiffFiles"))
				return err
			}
			_, err = f.Write([]byte(fileInfo.Content))
			if err != nil {
				Logger.Error(err.Error(), zap.String("method", "DiffFiles"))
				return err
			}
			Logger.Info(fmt.Sprintf("added file %s from resource name %s, resource UID %s and resource version %s",
				fileName, fileInfo.ResourceName, fileInfo.ResourceUID, fileInfo.ResourceVersion), zap.String("method", "DiffFiles"))
		} else if oldFiles[fileName] != fileInfo {
			f, err := os.Create(fileName)
			if err != nil {
				return err
			}
			_, err = f.Write([]byte(fileInfo.Content))
			if err != nil {
				return err
			}
			Logger.Info(fmt.Sprintf("modified file %s from resource name %s, resource UID %s from resoruce version %s to resource version %s",
				fileName, fileInfo.ResourceName, fileInfo.ResourceUID, oldFiles[fileName].ResourceVersion, fileInfo.ResourceVersion), zap.String("method", "DiffFiles"))
		}
	}
	for fileName, fileInfo := range oldFiles {
		if newFiles[fileName] == nilFileInfo {
			os.Remove(fileName)
			Logger.Info(fmt.Sprintf("deleted file %s from resource name %s, resource UID %s and resource version %s",
				fileName, fileInfo.ResourceName, fileInfo.ResourceUID, fileInfo.ResourceVersion), zap.String("method", "DiffFiles"))
		}
	}
	return nil
}
