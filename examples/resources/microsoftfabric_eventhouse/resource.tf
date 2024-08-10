resource "microsoftfabric_eventhouse" "example_eventhouse" {
  workspace_id = microsoftfabric_workspace.example.id
  display_name = "example_eventhouse_demo"
  description  = "An example_eventhouse description"
}
