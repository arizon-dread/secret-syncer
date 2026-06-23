package models

type Config struct {
	KubeAPI          KubeAPI         `yaml:"kube-api"`
	SecretServer     SecretServerAPI `yaml:"secret-server"`
	MonitoredSecrets []KubeSecret    `yaml:"monitored-secrets"`
}
type SecretServerAPI struct {
	TokenURL string `yaml:"tokenUrl"`
	BaseURL  string `yaml:"baseUrl"`
}
type SecretServerEntry struct {
	ServiceAccount        string                 `yaml:"serviceAccount"`
	Password              string                 `yaml:"password"`
	GrantType             string                 `yaml:"grantType"`
	SecretURLPath         string                 `yaml:"secretUrlPath"`
	FieldPropertyMappings []FieldPropertyMapping `yaml:"fieldPropertyMappings`
}
type FieldPropertyMapping struct {
	FieldName              string `yaml:"fieldName"`
	KubeSecretPropertyName string `yaml:"kubeSecretPropertyName"`
}
type KubeAPI struct {
	ServiceAccount string `yaml:"serviceAccount"`
	URL            string `yaml:"url"`
}
type KubeSecret struct {
	Name                 string              `yaml:"name"`
	KubernetesSecretName string              `yaml:"kubeSecretName"`
	SecretServerEntry    []SecretServerEntry `yaml:"secretServer"`
}
