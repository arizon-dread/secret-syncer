# secret-syncer

A go app, meant to be run as a cronjob in kubernetes that syncs secrets from SecretServer to kubernetes secrets

## Configuration

Example yaml:

```yaml
kube-api:
  serviceAccount: default
  url: https://kubernetes.default.svc
secret-server:
  tokenURL: https://secret-server.local/api/token
  baseURL: https://secret-server.local/api/v1
monitored-secrets:
  - name: "creds"
    kubeSecretName: "credentials"
    secretServer:
      - serviceAccount: "secret-server-account-name"
        password: "S3cr3tP455w0rd" # Can be set with the env var MONITORED_SECRETS_CREDS_SECRET_SERVER_PASSWORD in this example
        secretUrlPath: "/secret/123"
        fieldPropertyMappings:
          - fieldName: "jsonpathentry.nextentry.password" # The JSON structure down to the field you want to retrieve from SecretServer
            kubeSecretPropertyName: "DATABASE_PASSWORD" # The property name in the kubernetes secret data field.
    
```
