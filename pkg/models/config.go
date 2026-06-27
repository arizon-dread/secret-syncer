// Package models contains structs for the application
package models

type Config struct {
	KubeAPI          KubeAPI         `mapstructure:"kube-api"`
	SecretServer     SecretServerAPI `mapstructure:"secret-server"`
	MonitoredSecrets []KubeSecret    `mapstructure:"monitored-secrets"`
}
type SecretServerAPI struct {
	TokenURL string `mapstructure:"tokenUrl"`
	BaseURL  string `mapstructure:"baseUrl"`
}
type SecretServerEntry struct {
	ServiceAccount        string                 `mapstructure:"serviceAccount"`
	Password              string                 `mapstructure:"password"`
	GrantType             string                 `mapstructure:"grantType"`
	SecretURLPath         string                 `mapstructure:"secretUrlPath"`
	FieldPropertyMappings []FieldPropertyMapping `mapstructure:"fieldPropertyMappings"`
}
type FieldPropertyMapping struct {
	FieldName              string `mapstructure:"fieldName"`
	KubeSecretPropertyName string `mapstructure:"kubeSecretPropertyName"`
}
type KubeAPI struct {
	ServiceAccount string `mapstructure:"serviceAccount"`
	URL            string `mapstructure:"url"`
}
type KubeSecret struct {
	Name                 string              `mapstructure:"name"`
	KubernetesSecretName string              `mapstructure:"kubeSecretName"`
	SecretServerEntry    []SecretServerEntry `mapstructure:"secretServer"`
}
