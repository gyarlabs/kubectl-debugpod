package root

import (
	"fmt"
	"os"
	"strings"

	"github.com/gyarlabs/kubectl-debugpod/internal/cluster"
	"github.com/gyarlabs/kubectl-debugpod/internal/debugpod"
	"github.com/gyarlabs/kubectl-debugpod/internal/limits"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespace    string
	nodeName     string
	image        string
	stay         bool
	useBash      bool
	clusterCheck bool
	clusterArgs  []string
	checkLimits  bool
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-debugpod",
	Short: "A Kubernetes CLI plugin to launch debug pods and check cluster state",
	RunE: func(cmd *cobra.Command, args []string) error {
		if clusterCheck {
			cluster.RunClusterCheck(clusterArgs)
			return nil
		}

		if checkLimits {
			return limits.CheckLimits(namespace)
		}

		if strings.EqualFold(namespace, "all") {
			return fmt.Errorf("--namespace all is only supported together with --check-limits")
		}

		debugpod.RunDebugPod(debugpod.DebugOptions{
			Namespace:      namespace,
			NodeName:       nodeName,
			Image:          image,
			Stay:           stay,
			UseBash:        useBash,
			ServiceAccount: "",
			ClusterCheck:   clusterCheck,
		})
		return nil
	},
}

func init() {
	defaultNamespace := detectDefaultNamespace()

	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", defaultNamespace, "Target namespace (defaults to current kubectl namespace; use 'all' with --check-limits)")
	rootCmd.PersistentFlags().StringVar(&nodeName, "node", "", "Schedule the pod to a specific node")
	rootCmd.PersistentFlags().StringVar(&image, "image", "arsaphone/debugpod:v2", "Docker image for the debug pod")
	rootCmd.PersistentFlags().BoolVar(&stay, "stay", false, "Don't auto-delete the pod after exit")
	rootCmd.PersistentFlags().BoolVar(&useBash, "bash", false, "Use /bin/bash as the shell")

	rootCmd.PersistentFlags().BoolVar(&clusterCheck, "cluster-check", false, "Run k8sgpt analysis")
	rootCmd.PersistentFlags().StringSliceVar(&clusterArgs, "cluster-args", []string{}, "Extra arguments for k8sgpt")

	rootCmd.PersistentFlags().BoolVar(&checkLimits, "check-limits", false, "Check for missing resource limits in deployments")

	rootCmd.MarkFlagsMutuallyExclusive("cluster-check", "check-limits")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func detectDefaultNamespace() string {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	ns, _, err := config.Namespace()
	if err == nil && ns != "" {
		return ns
	}

	return "default"
}
