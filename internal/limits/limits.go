package limits

import (
    "context"
    "fmt"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

func CheckLimits() error {
    config, err := rest.InClusterConfig()
    if err != nil {
        return fmt.Errorf("failed to get in-cluster config: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return fmt.Errorf("failed to create clientset: %w", err)
    }

    namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list namespaces: %w", err)
    }

    for _, ns := range namespaces.Items {
        deployments, err := clientset.AppsV1().Deployments(ns.Name).List(context.TODO(), metav1.ListOptions{})
        if err != nil {
            return fmt.Errorf("failed to list deployments in namespace %s: %w", ns.Name, err)
        }

        for _, deploy := range deployments.Items {
            for _, container := range deploy.Spec.Template.Spec.Containers {
                if container.Resources.Limits == nil {
                    fmt.Printf("Deployment %s in namespace %s has container %s without resource limits\n",
                        deploy.Name, ns.Name, container.Name)
                }
            }
        }
    }

    return nil
}
