package root

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/gyarlabs/kubectl-debugpod/internal/cluster"
	"github.com/gyarlabs/kubectl-debugpod/internal/debugpod"
	"github.com/gyarlabs/kubectl-debugpod/internal/limits"
	"github.com/gyarlabs/kubectl-debugpod/internal/secrets"
)

var (
	namespace     string
	nodeName      string
	image         string
	stay          bool
	useBash       bool
	clusterCheck  bool
	clusterArgs   []string
	secretsNs     string
	secretsName   string
	decodeSecrets bool
	checkLimits   bool
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-debugpod",
	Short: "A Kubernetes CLI plugin to launch debug pods and check cluster state",
	Run: func(cmd *cobra.Command, args []string) {
		if clusterCheck {
			cluster.RunClusterCheck(clusterArgs)
			return
		}

		if secretsNs != "" {
			secrets.GetSecrets(secretsNs, secretsName, decodeSecrets)
			return
		}

		if checkLimits {
			limits.CheckLimits()
			return
		}

		debugpod.RunDebugPod(debugpod.DebugOptions{
	Namespace: namespace,
	NodeName:  nodeName,
	Image:     image,
	Stay:      stay,
	UseBash:   useBash,
})
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Target namespace")
	rootCmd.PersistentFlags().StringVar(&nodeName, "node", "", "Schedule the pod to a specific node")
	rootCmd.PersistentFlags().StringVar(&image, "image", "arsaphone/debugpod:v2", "Docker image for the debug pod")
	rootCmd.PersistentFlags().BoolVar(&stay, "stay", false, "Don't auto-delete the pod after exit")
	rootCmd.PersistentFlags().BoolVar(&useBash, "bash", false, "Use /bin/bash as the shell")

	rootCmd.PersistentFlags().BoolVar(&clusterCheck, "cluster-check", false, "Run k8sgpt analysis")
	rootCmd.PersistentFlags().StringSliceVar(&clusterArgs, "cluster-args", []string{}, "Extra arguments for k8sgpt")

	rootCmd.PersistentFlags().StringVar(&secretsNs, "secrets", "", "Fetch secrets from a specific namespace")
	rootCmd.PersistentFlags().StringVar(&secretsName, "secret-name", "", "Specific secret name to filter")
	rootCmd.PersistentFlags().BoolVar(&decodeSecrets, "decode", false, "Decode secret values (base64)")

	rootCmd.PersistentFlags().BoolVar(&checkLimits, "check-limits", false, "Check for missing resource limits in deployments")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
