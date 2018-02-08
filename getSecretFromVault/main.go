package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//extv1beta1 "k8s.io/api/extensions/v1beta1"
	"flag"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"strings"
	"fmt"
)

func main() {
	cn := os.Getenv("CERT_NAME")
	role := os.Getenv("ROLE_NAME")
	cert, key, err := vaultPKI(cn,role)
	if err != nil {
		panic(err.Error())
	}
	kubernetesSecret(cn, cert, key)
}
func vaultPKI(cn, role string) (cert string, key string, err error) {
	client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	token := os.Getenv("VAULT_TOKEN")
	client.SetToken(token)
	c := client.Logical()

	secret, err := c.Write("pki/issue/"+role,
		map[string]interface{}{
			"common_name": cn,
		})
	if err != nil {
		panic(err.Error())
	}
	data := secret.Data
	if secret == nil {
		return "none", "none", fmt.Errorf("Returned secret was nil")
	}

	if err != nil {
		return "none", "none", fmt.Errorf("Error parsing secret: %s", err)
	}
	cert = data["certificate"].(string)
	key = data["private_key"].(string)
	log.Println("cert and key is:", cert, key)
	if err != nil {
		return "none", "none", fmt.Errorf("Could not get TLS config: %s", err)
	}

	return cert, key, nil
}

func kubernetesSecret(name, cert, key string) error {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	var config *rest.Config
	var err error
	var kubeconfig *string
	if len(host) == 0 || len(port) == 0 {
		log.Println("unable to load in-cluster configuration, trying out cluster config")
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// creates the in-cluster config
		log.Println("Tryinh in lcuster config cluster config")
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	//creates the clientset

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	log.Println("cert is: ", cert)
	log.Println("key is: ", key)
	if err != nil {
		panic(err.Error())
	}
	secretName := strings.ToLower(name)
	log.Println("Saving secret to ", name)
	secretData := map[string][]byte{
		"tls.key": []byte(key),
		"tls.crt": []byte(cert),
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: secretData,
		Type: corev1.SecretTypeOpaque,
	}
	s, err := clientset.CoreV1().Secrets("default").Get(secret.Name, metav1.GetOptions{})
	if err != nil {
		log.Println("Creating secret")
		s, err = clientset.CoreV1().Secrets("default").Create(secret)
	} else {
		log.Println("Updating secret")
		s, err = clientset.CoreV1().Secrets("default").Update(secret)
	}
	if err != nil {
		panic(err.Error())
	}
	log.Printf("Secret %s  created in namespace %s\n", s.Name, secret.Namespace)
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
