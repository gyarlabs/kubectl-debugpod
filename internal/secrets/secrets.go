package secrets

import (
    "context"
    "encoding/base64"
    "fmt"
    "os"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

func GetSecrets(namespace, name string, decode bool) error {
    config, err := rest.InClusterConfig()
    if err != nil {
        return fmt.Errorf("failed to get in-cluster config: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return fmt.Errorf("failed to create clientset: %w", err)
    }

    var secrets []string

    if name == "" {
        secretList, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
        if err != nil {
            return fmt.Errorf("failed to list secrets: %w", err)
        }

        for _, s := range secretList.Items {
            secrets = append(secrets, s.Name)
        }
    } else {
        secrets = append(secrets, name)
    }

    for _, secName := range secrets {
        secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secName, metav1.GetOptions{})
        if err != nil {
            fmt.Fprintf(os.Stderr, "failed to get secret %s: %v\n", secName, err)
            continue
        }

        fmt.Printf("Secret: %s\n", secName)
        for k, v := range secret.Data {
            if decode {
                fmt.Printf("  %s: %s\n", k, string(v))
            } else {
                fmt.Printf("  %s: %s\n", k, base64.StdEncoding.EncodeToString(v))
            }
        }
    }

    return nil
}
