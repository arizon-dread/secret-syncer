package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

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
		return fmt.Errorf("unable to get config, %v", err)
	}
	noOfGoRoutines := len(config.MonitoredSecrets)
	ch := make(chan models.Result)
	for _, v := range config.MonitoredSecrets {
		go updateKubeSecret(v, ch)
	}
	for i := 0; i < noOfGoRoutines; i++ {
		res := <-ch
		if res.Err != nil {
			return res.Err
		}
	}
	return nil
}

func updateKubeSecret(kubeSecret models.KubeSecret, ch chan models.Result) {
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
		kSecret.ObjectMeta = metav1.ObjectMeta{
			Name:      kubeSecret.KubernetesSecretName,
			Namespace: namespace,
		}
		kSecret, err = clientSet.CoreV1().Secrets(namespace).Create(context.TODO(), kSecret, metav1.CreateOptions{})
		if err != nil {
			ch <- models.Result{Err: fmt.Errorf("unable to create secret, quitting, err : %v", err)}
		}
	}

	for _, s := range kubeSecret.SecretServerEntry {

		ssResp, err := getSecretServerSecret(s)
		if err != nil {
			log.Printf("error getting secret from secret server, trying next in config")
			continue
		}
		doSecretMapping(ssResp, s, kSecret, kubeSecret)
		ch <- models.Result{Err: nil}
	}
	_, err = clientSet.CoreV1().Secrets(namespace).Update(context.TODO(), kSecret, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("error updating secret %v, err. %v", kubeSecret.KubernetesSecretName, err)
		return
	}
	log.Printf("updated secret %v successfully", kubeSecret.KubernetesSecretName)
}

func getSecretServerSecret(ssSecret models.SecretServerEntry) (*models.SecretServerResponse, error) {
	token, err := getToken(ssSecret)
	if err != nil {
		log.Printf("failed to get token from SecretServer, %v", err)
		return nil, err
	}
	client := &http.Client{}
	path, err := url.JoinPath(config.SecretServer.BaseURL, ssSecret.SecretURLPath)
	if err != nil {
		log.Printf("unable to create a url path based on baseURL and SecretURLPath")
	}
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.Printf("failed to create request, %v", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("unable to get secret from %v, err: %v", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("unable to read body, %v", err)
		return nil, err
	}
	var ssResp *models.SecretServerResponse
	err = json.Unmarshal(body, &ssResp)
	if err != nil {
		log.Printf("unable to unmarshal response from secret server, %v", err)
		return nil, err
	}
	if ssResp.Message != "" {
		log.Printf("%v", ssResp.Message)
	}
	return ssResp, nil
}

func doSecretMapping(ssResp *models.SecretServerResponse, ssSecret models.SecretServerEntry, kSecret *v1.Secret, kubeSecret models.KubeSecret) {
	for _, v := range ssSecret.FieldPropertyMappings {
		if kSecret.Data == nil {
			kSecret.Data = make(map[string][]byte)
		}
		secretValue, exists := kSecret.Data[v.KubeSecretPropertyName]
		var err error
		if exists {
			if err != nil {
				log.Printf("failed to retrieve secret value, will create it %v", err)
			}
		}
		var ssValue string
		for _, item := range ssResp.Items {
			if item.FieldName == v.FieldName {
				ssValue = item.ItemValue
			}
		}
		if string(secretValue) == ssValue && len(secretValue) > 0 {
			log.Printf("secret %v property %v is up-to-date", kubeSecret.KubernetesSecretName, v.KubeSecretPropertyName)
		} else if ssValue != "" {
			// src := []byte(ssValue)
			// bVal := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
			// base64.StdEncoding.Encode(bVal, src)
			// kSecret.Data[v.KubeSecretPropertyName] = bVal
			kSecret.Data[v.KubeSecretPropertyName] = []byte(ssValue)
			log.Printf("update secret %v property %v", kubeSecret.KubernetesSecretName, v.KubeSecretPropertyName)
		} else {
			log.Printf("the value in SecretServer seems to be empty, will not overwrite kubernetes secret")
		}
	}
}

func getToken(ssSecret models.SecretServerEntry) (string, error) {
	client := &http.Client{}
	form := url.Values{}
	form.Add("username", ssSecret.ServiceAccount)
	form.Add("password", ssSecret.Password)
	form.Add("grant_type", ssSecret.GrantType)

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
	if len(token) > 0 {
		log.Printf("access_token was successfully retrieved")
	}
	return token, nil
}
