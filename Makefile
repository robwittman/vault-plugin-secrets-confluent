build:
	go build -o vault/plugins/vault-plugin-secrets-confluent cmd/vault-plugin-secrets-confluent/main.go

vault-dev:
	vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins -log-level=debug

setup:
	vault plugin info secret vault-plugin-secrets-confluent
	vault secrets enable -path=confluent vault-plugin-secrets-confluent
