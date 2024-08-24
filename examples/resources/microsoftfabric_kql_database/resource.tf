resource "microsoftfabric_kqldatabase" "example_kql_database" {
  workspace_id = microsoftfabric_workspace.example.id
  display_name = "example_kql_database_demo"
  description = "I am a description."
  creation_payload = {
    database_type = "ReadWrite"
    parent_eventhouse_items_id = microsoftfabric_eventhouse.example_eventhouse.id
  }
}