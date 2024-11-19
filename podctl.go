package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/charmbracelet/huh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func main() {
	// Define color codes for terminal output
	colorReset := "\033[0m"
	colorCyan := "\033[36m"
	colorGreen := "\033[32m"
	maxHeight := 15

	// Load kubeconfig file
	kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// Get available contexts from kubeconfig
	contexts := getContexts(config)
	height := len(contexts)
	if height > maxHeight {
		height = maxHeight
	}

	// Prompt user to select a Kubernetes context
	contextResult := ""
	contextPrompt := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Cluster").
				Options(huh.NewOptions(contexts...)...).
				Value(&contextResult).
				WithHeight(height + 2)))

	contextError := contextPrompt.Run()
	if contextError != nil {
		errorHandle(contextError)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s %s %s %s %s \n", colorGreen, "\uf00c", colorCyan, contextResult, colorReset)

	// Create Kubernetes client configuration for the selected context
	clientConfig := clientcmd.NewNonInteractiveClientConfig(*config, contextResult, &clientcmd.ConfigOverrides{}, nil)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		log.Fatalf("Failed to get client config: %v", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// List namespaces in the selected context
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list namespaces: %v", err)
	}

	// Collect namespace names
	var namespaceNames []string
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.Name)
	}

	// Prompt user to select a namespace
	namespaceResult := ""
	namespacePrompt := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Namespace").
				Options(huh.NewOptions(namespaceNames...)...).
				Value(&namespaceResult).
				WithHeight(maxHeight)))

	namespaceError := namespacePrompt.Run()
	if namespaceError != nil {
		errorHandle(namespaceError)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s %s %s %s %s \n", colorGreen, "\uf00c", colorCyan, namespaceResult, colorReset)

	// List pods in the selected namespace
	pods, err := clientset.CoreV1().Pods(namespaceResult).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list pods: %v", err)
	}

	// Collect pod names
	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	// Prompt user to select a pod
	podResult := ""
	podPrompt := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pod").
				Options(huh.NewOptions(podNames...)...).
				Value(&podResult).
				WithHeight(maxHeight)))

	podError := podPrompt.Run()
	if podError != nil {
		errorHandle(podError)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s %s %s %s %s \n", colorGreen, "\uf00c", colorCyan, podResult, colorReset)

	// Get the selected pod details
	selectedPod, err := clientset.CoreV1().Pods(namespaceResult).Get(context.TODO(), podResult, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get selected pod: %v", err)
	}

	// Collect container names from the selected pod
	var containerNames []string
	for _, container := range selectedPod.Spec.Containers {
		containerNames = append(containerNames, container.Name)
	}

	// Prompt user to select a container
	containerResult := ""
	containerPrompt := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Container").
				Options(huh.NewOptions(containerNames...)...).
				Value(&containerResult).
				WithHeight(len(containerNames) + 2)))

	containerError := containerPrompt.Run()
	if containerError != nil {
		errorHandle(containerError)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s %s %s %s %s \n", colorGreen, "\uf00c", colorCyan, containerResult, colorReset)

	// Execute a shell in the selected container
	cmd := exec.Command("kubectl", "exec", "-it", podResult, "-c", containerResult, "--context="+contextResult, "-n", namespaceResult, "--", "sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		if err == io.EOF {
			fmt.Println("Exiting shell (EOF detected).")
		} else {
			log.Fatalf("Command execution failed: %v", err)
		}
	}

	fmt.Printf("Exited out of \"%s / %s\"", podResult, containerResult)
}

// errorHandle handles errors from user input prompts
func errorHandle(err error) {
	if err.Error() == "user aborted" {
		fmt.Println("Process was cancelled by the user.")
	} else {
		fmt.Println("Selection failed:", err)
	}
}

// getContexts retrieves the list of contexts from the kubeconfig
func getContexts(config *api.Config) []string {
	var contexts []string
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}
	return contexts
}
