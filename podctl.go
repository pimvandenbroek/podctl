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

	// Get contexts from kubeconfig
	contexts := getContexts(config)
	// Set the height limit for the context prompt
	height := len(contexts)
	if height > maxHeight {
		height = maxHeight
	}

	contextResult := ""
	contextPrompt := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().
		Title("Cluster").
		Options(huh.NewOptions(contexts...)...).
		Value(&contextResult).
		WithHeight(height + 2)))

	contextError := contextPrompt.Run()
	if contextError != nil {
		if contextError.Error() == "user aborted" {
			// Handle Ctrl+C (interruption)
			fmt.Println("Process was cancelled by the user.")
		} else {
			// Handle other errors
			fmt.Println("Selection failed:", contextError)
		}
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s %s %s %s %s \n", colorGreen, "\uf00c", colorCyan, contextResult, colorReset)

	// Switch to selected context
	clientConfig := clientcmd.NewNonInteractiveClientConfig(*config, contextResult, &clientcmd.ConfigOverrides{}, nil)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		log.Fatalf("Failed to get client config: %v", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// List namespaces
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list namespaces: %v", err)
	}

	// Get namespace names
	var namespaceNames []string
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.Name)
	}

	namespaceResult := ""
	namespacePrompt := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().
		Title("Namespace").
		Options(huh.NewOptions(namespaceNames...)...).
		Value(&namespaceResult).
		WithHeight(maxHeight)))

	namespaceError := namespacePrompt.Run()
	if namespaceError != nil {
		if namespaceError.Error() == "user aborted" {
			// Handle Ctrl+C (interruption)
			fmt.Println("Process was cancelled by the user.")
		} else {
			// Handle other errors
			fmt.Println("Selection failed:", namespaceError)
		}
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s %s %s %s %s \n", colorGreen, "\uf00c", colorCyan, namespaceResult, colorReset)

	// List pods in the selected namespace
	pods, err := clientset.CoreV1().Pods(namespaceResult).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list pods: %v", err)
	}

	// Get pod names
	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	// Create pod prompt
	podResult := ""
	podPrompt := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().
		Title("Pod").
		Options(huh.NewOptions(podNames...)...).
		Value(&podResult).
		WithHeight(maxHeight)))

	podError := podPrompt.Run()
	if podError != nil {
		if podError.Error() == "user aborted" {
			// Handle Ctrl+C (interruption)
			fmt.Println("Process was cancelled by the user.")
		} else {
			// Handle other errors
			fmt.Println("Selection failed:", podError)
		}
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s %s %s %s %s \n", colorGreen, "\uf00c", colorCyan, podResult, colorReset)

	// List containers in the selected pod
	selectedPod, err := clientset.CoreV1().Pods(namespaceResult).Get(context.TODO(), podResult, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get selected pod: %v", err)
	}

	var containerNames []string
	for _, container := range selectedPod.Spec.Containers {
		containerNames = append(containerNames, container.Name)
	}

	// Create container prompt
	containerResult := ""
	containerPrompt := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().
		Title("Container").
		Options(huh.NewOptions(containerNames...)...).
		Value(&containerResult).
		WithHeight(len(containerNames) + 2)))

	containerError := containerPrompt.Run()
	if containerError != nil {
		if containerError.Error() == "user aborted" {
			// Handle Ctrl+C (interruption)
			fmt.Println("Process was cancelled by the user.")
		} else {
			// Handle other errors
			fmt.Println("Selection failed:", containerError)
		}
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%s %s %s %s %s \n", colorGreen, "\uf00c", colorCyan, containerResult, colorReset)
	cmd := exec.Command("kubectl", "exec", "-it", podResult, "-c", containerResult, "--context="+contextResult, "-n", namespaceResult, "--", "sh")

	// Set the command's standard input/output to the current process's
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the kubectl command
	err = cmd.Run()

	// Handle EOF (CTRL+D) as expected behavior
	if err != nil {
		if err == io.EOF {
			// EOF (CTRL+D) is expected, so treat it as a success
			fmt.Println("Exiting shell (EOF detected).")
		} else {
			// If it's some other error, log it
			log.Fatalf("Command execution failed: %v", err)
		}
	}

	fmt.Printf("Exited out of \"%s / %s\"", podResult, containerResult)
}

// getContexts retrieves the context names from the kubeconfig
func getContexts(config *api.Config) []string {
	var contexts []string
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}
	return contexts
}
