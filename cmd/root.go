package cmd

import (
	"os"
	"sidecar/service"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	namespace     string
	resource      string
	sleepTime     int
	fileK8sConfig string
	fileConfig    string
	debug         bool
	Logger        *zap.Logger
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "sidecar",
	Short: "An application collect and watch configmap and secret in a kubernetes with a specified label.",
	Long:  `An application collect and watch configmap and secret in a kubernetes with a specified label.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		Logger, _ = zap.NewDevelopment(
			zap.AddStacktrace(zapcore.FatalLevel),
			zap.AddCallerSkip(0),
		)
		Logger = Logger.With(zap.String("service", "sidecar"))
		service.Logger = Logger.With(zap.String("package", "sidecar"))
		err := service.Init(sleepTime, fileK8sConfig, fileConfig)
		if err != nil {
			Logger.Error(err.Error())
		}
		return err
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		if debug == true {
			time.Sleep(1 * time.Minute)
		}
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().IntVarP(&sleepTime, "sleep-time", "T", 3, "How many seconds to next check")
	RootCmd.PersistentFlags().StringVarP(&fileK8sConfig, "kube-config", "k", "", "Where is the file kubeconfig? If you run the application in pod, please ingnore this flag.")
	RootCmd.PersistentFlags().StringVarP(&fileConfig, "config", "c", "", "file config to define resource to watch")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "if the flag debug is true then application will slepp 1 minute before exit when it have error.")
}
