# secret-syncer

A go app, meant to be run as a CronJob in Kubernetes that syncs secrets from Secret Server to Kubernetes secrets.  
Secret Server is treated as a single source of truth.  
This app imports secrets directly from the source of truth into the runtime environment so no secrets needs to be added to the VCS.

## Configuration

Example config yaml:

```yaml
kube-api:
  serviceAccount: syncer
  url: https://kubernetes.default.svc
secret-server:
  tokenURL: https://secret-server.local/oauth2/token
  baseURL: https://secret-server.local/api/v1
monitored-secrets:
  - name: "creds"
    kubeSecretName: "credentials"
    secretServer:
      - serviceAccount: "secret-server-account-name"
        password: "S3cr3tP455w0rd" # Can be set with the env var MONITORED_SECRETS_CREDS_SECRET_SERVER_PASSWORD in this example
        grantType: password # This is the grant_type for the token retrieval.
        secretUrlPath: "/secrets/123"
        fieldPropertyMappings:
          - fieldName: "Password" # The item.fieldName you want to retrieve the fieldValue from in the SecretServer secret
            kubeSecretPropertyName: "DATABASE_PASSWORD" # The property name in the kubernetes secret data field.
```

## Kubernetes manifests

Please see the [kubernetes](./kubernetes/) folder for an example deployment.  
Note: The serviceAccount needs to have a role and rolebinding that allows for it to read and update secrets in the namespace where it's running. It would be reasonable to not use the same serviceAccount for a pod that is exposed outside the namespace.

## DISCLAIMER

Use of this client requires a properly licensed Secret Server environment and is subject to the user’s agreement with Delinea.  
This application is a community project and is not in any way maintained or affiliated with Delinea, it is only implementing an integration with the Secret Server API and consumes the API such as having Go structs that align with the JSON response returned from the API when retrieving a secret.  
See the [LICENSE](./LICENSE) for more information.
