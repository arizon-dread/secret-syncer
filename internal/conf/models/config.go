package models

type Config struct {
	KubeAPI          KubeAPI      `yaml:"kube-api"`
	SecretServer     SecretServer `yaml:"secret-server"`
	MonitoredSecrets []KubeSecret `yaml:"monitored-secrets"`
}
type SecretServer struct {
	TokenURL string `yaml:"tokenUrl"`
	BaseURL  string `yaml:"baseUrl"`
}
type SecretServerSecret struct {
	ServiceAccount        string                 `yaml:"serviceAccount"`
	Password              string                 `yaml:"Password"`
	SecretURLPath         string                 `yaml:"secretUrlPath"`
	FieldPropertyMappings []FieldPropertyMapping `yaml:"fieldPropertyMappings`
}
type FieldPropertyMapping struct {
	FieldPath              string `yaml:"fieldPath"`
	KubeSecretPropertyName string `yaml:"kubeSecretPropertyName"`
}
type KubeAPI struct {
	ServiceAccount string `yaml:"serviceAccount"`
	URL            string `yaml:"url"`
}
type KubeSecret struct {
	KubernetesSecretName string               `yaml:"kubeSecretName"`
	SecretServerSecret   []SecretServerSecret `yaml:"secretServer"`
}
