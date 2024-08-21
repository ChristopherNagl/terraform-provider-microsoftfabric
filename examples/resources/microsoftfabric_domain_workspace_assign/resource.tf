resource "microsoftfabric_domain_workspace_assign" "example_domain_assignment" {
  domain_id = "domain-id"
  workspace_ids = [
    microsoftfabric_workspace.example.id,
    "workspace-id"
  ]
}