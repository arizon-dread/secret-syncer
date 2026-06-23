# secret-syncer

A go app, meant to be run as a cronjob in kubernetes that syncs secrets from SecretServer to kubernetes secrets

## Configuration

Example config yaml:

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
        grantType: password # This is the grant_type for the token retrieval.
        secretUrlPath: "/secret/123"
        fieldPropertyMappings:
          - fieldName: "Password" # The item.fieldName you want to retrieve the fieldValue from in the SecretServer secret
            kubeSecretPropertyName: "DATABASE_PASSWORD" # The property name in the kubernetes secret data field.
```
