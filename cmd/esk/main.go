package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
)

var rootCmd = &cobra.Command{
	Use:   "esk",
	Short: "ESK - Edge Service for Kubernetes",
	Long:  `ESK is a Kubernetes management tool for edge computing scenarios.`,
}

var clientset *kubernetes.Clientset

func init() {
	cobra.OnInitialize(initKubeClient)

	// Add commands
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(getCmd())
}

func initKubeClient() {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating kubernetes client: %v\n", err)
		os.Exit(1)
	}
}

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [resource]",
		Short: "List Kubernetes resources",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Resource type required")
				return
			}

			switch args[0] {
			case "pod", "pods":
				listPods(cmd)
			default:
				fmt.Printf("Unknown resource type: %s\n", args[0])
			}
		},
	}
	return cmd
}

func getCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [resource]",
		Short: "Get Kubernetes resource details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Resource type required")
				return
			}

			switch args[0] {
			case "deployment", "deployments":
				getDeployments(cmd)
			default:
				fmt.Printf("Unknown resource type: %s\n", args[0])
			}
		},
	}
	return cmd
}

func listPods(cmd *cobra.Command) {
	pods, err := clientset.CoreV1().Pods("").List(cmd.Context(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		return
	}

	fmt.Printf("Found %d pods\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Printf("Pod: %s\tNamespace: %s\tStatus: %s\n",
			pod.Name, pod.Namespace, pod.Status.Phase)
	}
}

func getDeployments(cmd *cobra.Command) {
	deployments, err := clientset.AppsV1().Deployments("").List(cmd.Context(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing deployments: %v\n", err)
		return
	}

	fmt.Printf("Found %d deployments\n", len(deployments.Items))
	for _, deployment := range deployments.Items {
		fmt.Printf("Deployment: %s\tNamespace: %s\tReplicas: %d/%d\n",
			deployment.Name, deployment.Namespace,
			deployment.Status.ReadyReplicas, deployment.Status.Replicas)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
