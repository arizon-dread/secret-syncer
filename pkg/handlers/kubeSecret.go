package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/arizon-dread/secret-syncer/internal/conf"
	"github.com/arizon-dread/secret-syncer/pkg/models"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	config    *models.Config
	clientSet *kubernetes.Clientset
	namespace string
)

func SyncMonitoredSecrets() error {
	var err error
	config, err = conf.GetConfig()
	if err != nil {
		log.Fatalf("unable to get config, %v", err)
	}
	wg := &sync.WaitGroup{}
	ch := make(chan models.Result)
	noOfGoRoutines := 0
	for _, v := range config.MonitoredSecrets {
		wg.Add(1)
		noOfGoRoutines++
		go updateKubeSecret(v, ch, wg)
	}

	for res := range ch {
		if res.Err != nil {
			close(ch)
			return res.Err
		}
	}
	wg.Wait()
	return nil
}

func updateKubeSecret(kubeSecret models.KubeSecret, ch chan models.Result, wg *sync.WaitGroup) {
	defer wg.Done()
	clusterConf, err := rest.InClusterConfig()
	if err != nil {
		ch <- models.Result{Err: fmt.Errorf("failed to get cluster config, will not be able to see or touch secrets, quitting, err: %v", err)}
	}
	clientSet, err = kubernetes.NewForConfig(clusterConf)
	if err != nil {
		ch <- models.Result{Err: fmt.Errorf("failed to create kubernetes client, will not be able to see or touch secrets, quitting, err: %v", err)}
	}
	namespace = os.Getenv("NAMESPACE")
	kSecret, err := clientSet.CoreV1().Secrets(namespace).Get(context.TODO(), kubeSecret.KubernetesSecretName, metav1.GetOptions{})
	if err != nil {
		log.Printf("unable to get secret %v, will create it", kubeSecret.KubernetesSecretName)
		kSecret = &v1.Secret{}
		kSecret.Name = kubeSecret.KubernetesSecretName
		kSecret, err = clientSet.CoreV1().Secrets(namespace).Create(context.TODO(), kSecret, metav1.CreateOptions{})
		if err != nil {
			ch <- models.Result{Err: fmt.Errorf("unable to create secret, quitting, err : %v", err)}
		}
	}

	for _, s := range kubeSecret.SecretServerEntry {

		secretJSON, err := getSecretServerSecret(s)
		if err != nil {
			log.Printf("error getting secret from secret server, trying next in config")
			continue
		}
		doSecretMapping(secretJSON, s, kSecret, kubeSecret)
		ch <- models.Result{Err: nil}
	}
}

func getSecretServerSecret(ssSecret models.SecretServerEntry) (string, error) {
	token, err := getToken(ssSecret)
	if err != nil {
		log.Printf("failed to get token from SecretServer, %v", err)
		return "", err
	}
	client := &http.Client{}
	path := path.Join(config.SecretServer.BaseURL, ssSecret.SecretURLPath)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.Printf("failed to create request, %v", err)
	}
	req.Header.Add("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("unable to get secret from %v, err: %v", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("unable to read body, %v", err)
		return "", err
	}
	return string(body), nil
}

func doSecretMapping(secretJSON string, ssSecret models.SecretServerEntry, kSecret *v1.Secret, kubeSecret models.KubeSecret) {
	var m map[string]any
	err := json.Unmarshal([]byte(secretJSON), &m)
	if err != nil {
		log.Printf("unable to marshal secret server response json into generic map, %v", err)
		return
	}
	for _, v := range ssSecret.FieldPropertyMappings {
		secretValue, exists := kSecret.Data[v.KubeSecretPropertyName]
		if exists {
			decodedValue, err := base64.StdEncoding.DecodeString(string(secretValue))
			if err != nil {
				log.Printf("Failed to decode secret value, err", err)
				continue
			}
			ssValue, exists := m[v.FieldPath].(string)
			if exists {
				if string(decodedValue) == ssValue {
					log.Printf("secret %v property %v is up-to-date", kubeSecret.KubernetesSecretName, v.KubeSecretPropertyName)
				} else {
					kSecret.Data[v.KubeSecretPropertyName] = []byte(ssValue)
					_, err = clientSet.CoreV1().Secrets(namespace).Update(context.TODO(), kSecret, metav1.UpdateOptions{})
					if err != nil {
						log.Printf("error updating secret %v, err. %v", kubeSecret.KubernetesSecretName, err)
						return
					}
					log.Printf("updated secret %v successfully", kubeSecret.KubernetesSecretName)
				}
			}
		}
	}
}

func getToken(ssSecret models.SecretServerEntry) (string, error) {
	client := &http.Client{}
	form := url.Values{}
	form.Add("user", ssSecret.ServiceAccount)
	form.Add("password", ssSecret.Password)
	// There's another value that needs to be set

	req, err := http.NewRequest("POST", config.SecretServer.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		log.Printf("error creating token request, %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error calling the token endpoint, %v", err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body, %v", err)
		return "", err
	}
	var m map[string]any
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Printf("error unmarshalling token response into generic go struct, %v", err)
	}
	token := m["access_token"].(string)
	return token, nil
}
