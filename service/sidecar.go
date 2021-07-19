package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"sidecar/models"
	"sidecar/service/configmap"
	"sidecar/service/secret"
	"time"

	"log"

	validator "github.com/go-playground/validator/v10"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Init(namespace string, sleepTime int, fileK8sConfig string, fileConfig string) error {
	yamlFile, err := ioutil.ReadFile(fileConfig)
	if err != nil {
		return err
	}
	var resources models.Resources
	err = yaml.Unmarshal(yamlFile, &resources)
	if err != nil {
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
		return err
	}
	clientK8s, err := kubernetes.NewForConfig(configK8s)
	if err != nil {
		return err
	}
	for _, resource := range resources.Resources {
		os.MkdirAll(resource.Path, 777)
	}
	return WatchResource(resources, clientK8s, namespace, sleepTime)
}

func WatchResource(resoures models.Resources, k8sClient *kubernetes.Clientset, namespace string, sleepTime int) error {
	fmt.Println(resoures)
	presentFiles := make(map[string]models.FileInfo)
	for {
		getFiles := make(map[string]models.FileInfo)
		for _, resource := range resoures.Resources {
			labelset := make(map[string]string)
			for _, label := range resource.Labels {
				labelset[label.Name] = label.Value
			}
			//fmt.Println(labelset)
			opt := v1.ListOptions{LabelSelector: labels.Set(labelset).String()}
			if resource.Type == "both" || resource.Type == "configmap" {
				err := configmap.WatchConfigMaps(k8sClient, namespace, opt, resource.Path, getFiles)
				if err != nil {
					return err
				}
			}
			if resource.Type == "both" || resource.Type == "secret" {
				err := secret.WatchSecrets(k8sClient, namespace, opt, resource.Path, getFiles)
				if err != nil {
					return err
				}
			}
			//fmt.Println(getFiles)
			err := DiffFiles(presentFiles, getFiles)
			if err != nil {
				return err
			}
			presentFiles = getFiles
		}
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return nil
}

func DiffFiles(oldFiles map[string]models.FileInfo, newFiles map[string]models.FileInfo) error {
	for fileName, fileInfo := range newFiles {
		nilFileInfo := models.FileInfo{}
		if oldFiles[fileName] == nilFileInfo {
			f, err := os.Create(fileName)
			if err != nil {
				return err
			}
			_, err = f.Write([]byte(fileInfo.Content))
			if err != nil {
				return err
			}
			log.Printf("added file %s from resource name %s, resource UID %s and resource version %s\n",
				fileName, fileInfo.ResourceName, fileInfo.ResourceUID, fileInfo.ResourceVersion)
		} else if oldFiles[fileName] != fileInfo {
			f, err := os.Create(fileName)
			if err != nil {
				return err
			}
			_, err = f.Write([]byte(fileInfo.Content))
			if err != nil {
				return err
			}
			log.Printf("modified file %s from resource name %s, resource UID %s from resoruce version %s to resource version %s\n",
				fileName, fileInfo.ResourceName, fileInfo.ResourceUID, oldFiles[fileName].ResourceVersion, fileInfo.ResourceVersion)
		}
	}
	return nil
}
