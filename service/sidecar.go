package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sidecar/models"
	"sidecar/service/configmap"
	"sidecar/service/secret"
	"sort"
	"time"

	validator "github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	nilFileInfo = models.FileInfo{}
	Logger      *zap.Logger
)

func Init(sleepTime int, fileK8sConfig string, fileConfig string) error {
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
	err = validate.Struct(&resources)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "Init"))
		return err
	}
	var configK8s *rest.Config
	if fileK8sConfig == "" {
		configK8s, err = rest.InClusterConfig()
	} else {
		configK8s, err = clientcmd.BuildConfigFromFlags("", fileK8sConfig)
	}

	Logger.Info(fmt.Sprintf("Config for cluster api at '%s' loaded.\n", configK8s.Host), zap.String("method", "Init"))
	clientK8s, err := kubernetes.NewForConfig(configK8s)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "Init"))
		return err
	}
	for i, resource := range resources.Resources {
		err := validate.Struct(&resources)
		if err != nil {
			Logger.Error(err.Error(), zap.String("method", "Init"))
			return err
		}
		err = os.MkdirAll(resource.Path, 0777)
		if err != nil {
			Logger.Warn(err.Error(), zap.String("method", "Init"))
		}
		resources.Resources[i].Index = i
	}
	//fmt.Println(resources.Resources)
	configmap.Logger = Logger.With(zap.String("package", "configmap"))
	secret.Logger = Logger.With(zap.String("package", "secret"))
	err = GetResource(&resources, clientK8s)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "Init"))
		return err
	}
	if len(resources.Resources) == 0 {
		return nil
	}
	return WatchResource(resources, clientK8s, sleepTime)
}
func GetResource(resoures *models.Resources, k8sClient *kubernetes.Clientset) error {
	sort.Slice(resoures.Resources, func(i, j int) bool {
		return resoures.Resources[i].Method < resoures.Resources[j].Method
	})
	getFiles := make(map[string]models.FileInfo)
	count := 0
	for _, resource := range resoures.Resources {
		if resource.Method != "get" {
			break
		}
		if resource.Type == "both" || resource.Type == "configmap" {
			err := configmap.ListConfigMaps(k8sClient, resource, getFiles)
			if err != nil {
				Logger.Warn(err.Error(), zap.String("method", "GetResource"))
				//return err
			}
		}
		if resource.Type == "both" || resource.Type == "secret" {
			err := secret.ListSecrets(k8sClient, resource, getFiles)
			if err != nil {
				Logger.Warn(err.Error(), zap.String("method", "GetResource"))
				//return err
			}
		}
		count++
	}
	for fileName, fileInfo := range getFiles {
		err := WriteFile(fileName, fileInfo.Content)
		if err != nil {
			Logger.Error(err.Error(), zap.String("method", "GetResource"))
			return err
		}
		Logger.Info(fmt.Sprintf("added file %s from namespace %s, resource name %s, resource UID %s and resource version %s",
			resoures.Resources[fileInfo.Index].Namespace, fileName, fileInfo.ResourceName, fileInfo.ResourceUID,
			fileInfo.ResourceVersion), zap.String("method", "GetResource"))
	}
	for i := 0; i < count; i++ {
		ExecInlines(resoures.Resources[i].ScriptInlines, i)
	}
	resoures.Resources = resoures.Resources[count:]
	return nil
}
func WatchResource(resoures models.Resources, k8sClient *kubernetes.Clientset, sleepTime int) error {
	presentFiles := make(map[string]models.FileInfo)
	for {
		getFiles := make(map[string]models.FileInfo)
		for _, resource := range resoures.Resources {

			if resource.Type == "both" || resource.Type == "configmap" {
				err := configmap.ListConfigMaps(k8sClient, resource, getFiles)
				if err != nil {
					Logger.Warn(err.Error(), zap.String("method", "WatchResource"))
					//return err
				}
			}
			if resource.Type == "both" || resource.Type == "secret" {
				err := secret.ListSecrets(k8sClient, resource, getFiles)
				if err != nil {
					Logger.Warn(err.Error(), zap.String("method", "WatchResource"))
					//return err
				}
			}
		}
		err := DiffFiles(resoures, presentFiles, getFiles)
		if err != nil {
			Logger.Error(err.Error(), zap.String("method", "WatchResource"))
			return err
		}
		presentFiles = getFiles
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return nil
}

func DiffFiles(resoures models.Resources, oldFiles map[string]models.FileInfo, newFiles map[string]models.FileInfo) error {
	checkResourceChange := make([]bool, len(resoures.Resources))
	for fileName, fileInfo := range newFiles {
		if oldFiles[fileName].ResourceName == "" {
			checkResourceChange[fileInfo.Index] = true
			err := WriteFile(fileName, fileInfo.Content)
			if err != nil {
				Logger.Error(err.Error(), zap.String("method", "DiffFiles"))
				return err
			}
			Logger.Info(fmt.Sprintf("added file %s from namespace %s, resource name %s, resource UID %s and resource version %s",
				resoures.Resources[fileInfo.Index].Namespace, fileName, fileInfo.ResourceName, fileInfo.ResourceUID, fileInfo.ResourceVersion), zap.String("method", "DiffFiles"))
		} else if oldFiles[fileName].ResourceUID != newFiles[fileName].ResourceUID || oldFiles[fileName].ResourceVersion != newFiles[fileName].ResourceVersion {
			checkResourceChange[fileInfo.Index] = true
			err := WriteFile(fileName, fileInfo.Content)
			if err != nil {
				Logger.Error(err.Error(), zap.String("method", "DiffFiles"))
				return err
			}
			Logger.Info(fmt.Sprintf("modified file %s from resource namepace %s, name %s, resource UID %s resoruce version %s to resource version %s",
				fileName, resoures.Resources[fileInfo.Index].Namespace, fileInfo.ResourceName, fileInfo.ResourceUID, oldFiles[fileName].ResourceVersion, fileInfo.ResourceVersion), zap.String("method", "DiffFiles"))
		}
	}
	for fileName, fileInfo := range oldFiles {
		if newFiles[fileName].ResourceName == "" {
			err := os.Remove(fileName)
			if err != nil {
				Logger.Error(err.Error(), zap.String("method", "DiffFiles"))
				return err
			}
			Logger.Info(fmt.Sprintf("deleted file %s from resource  namepace %s, name %s, resource UID %s and resource version %s",
				fileName, resoures.Resources[fileInfo.Index].Namespace, fileInfo.ResourceName, fileInfo.ResourceUID, fileInfo.ResourceVersion), zap.String("method", "DiffFiles"))
		}
	}
	for i, resource := range resoures.Resources {
		if checkResourceChange[i] {
			ExecInlines(resource.ScriptInlines, i)
		}
	}
	// fmt.Println(checkResourceChange)
	// fmt.Println(resoures.Resources)
	return nil
}
func WriteFile(fileName string, content []byte) error {
	f, err := os.Create(fileName)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "WriteFile"))
		return err
	}
	_, err = f.Write(content)
	if err != nil {
		Logger.Error(err.Error(), zap.String("method", "WriteFile"))
		return err
	}
	return nil
}

func ExecInlines(script []string, index int) {
	for i, command := range script {
		cmd := exec.Command("/bin/bash", "-c", command)
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err := cmd.Run()
		if err != nil {
			Logger.Warn(fmt.Sprintf("something wrong when exec inline %d script in resource %d with error %s and sdtderr:\n %s",
				i, index, err.Error(), errb.String()), zap.String("method", "ExecInlines"))
			continue
		}
		Logger.Info(fmt.Sprintf("sdtdout when exec inline %d script in resource %d:\n%s", i, index, outb.String()), zap.String("method", "ExecInlines"))
	}
}
