resource "microsoftfabric_lakehouse" "example_lakehouse" {
  workspace_id = microsoftfabric_workspace.example.id
  display_name = "lakehouse_demo"
  description  = "An example"
}