resource "microsoftfabric_workspace_capacity_assignment" "workspace_assignment" {
  workspace_id = microsoftfabric_workspace.example.id
  capacity_id  = "xxxxx"
}