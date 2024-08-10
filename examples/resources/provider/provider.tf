provider "microsoftfabric" {
  client_id       = "xxx"
  client_secret   = "Txxx"
  tenant_id       = "9xxx"
  token_file_path = "${path.module}/token.json"
}