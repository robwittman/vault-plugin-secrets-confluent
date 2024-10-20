# vault-plugin-secrets-confluent

## Quickstart 

#### Build the plugin and start Vault 
```shell 
make build 
make vault-dev
```

#### Enable the plugin 
```shell 
make setup
```

#### Configure the engine 
```shell 
vault write confluent/config \
  username='$API_KEY' \
  password='$API_SECRET' \
  url="https://api.confluent.cloud"
```

#### Create a role

**Note**: This role should map to an existing service account

```shell 
vault write confluent/role/test service_account="$SERVICE_ACCOUNT_ID"
```

#### Generate credentials 
```shell 
 vault read confluent/creds/test                                 
Key                Value
---                -----
lease_id           confluent/creds/test/hO1iPhNVJjLsCFyabUn3TmcI
lease_duration     768h
lease_renewable    true
api_key            <redacted>
api_secret         <redacted>
```

#### Test lease revocation 

This should delete the API Key within a few moments
```shell 
vault lease revoke confluent/creds/test/hO1iPhNVJjLsCFyabUn3TmcI
```