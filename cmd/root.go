package cmd

import (
	"log"
	"os"
	"sidecar/service"

	"github.com/spf13/cobra"
)

var (
	namespace     string
	resource      string
	sleepTime     int
	fileK8sConfig string
	fileConfig    string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "sidecar",
	Short: "An application collect and watch configmap and secret in a kubernetes with a specified label.",
	Long:  `An application collect and watch configmap and secret in a kubernetes with a specified label.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := service.Init(namespace, sleepTime, fileK8sConfig, fileConfig)
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&namespace, "namepspace", "n", "", "k8s namespace. If you watch on all namespace, please ingnore this flag.")
	RootCmd.PersistentFlags().StringVarP(&resource, "resource", "r", "configmap", "type of resource watching(configmap, secret or both)")
	RootCmd.PersistentFlags().IntVarP(&sleepTime, "sleep-time", "T", 3, "How many seconds to next check")
	RootCmd.PersistentFlags().StringVarP(&fileK8sConfig, "kube-config", "k", "", "Where is the file kubeconfig? If you run the application in pod, please ingnore this flag.")
	RootCmd.PersistentFlags().StringVarP(&fileConfig, "config", "c", "", "file config to define resource to watch")
}
