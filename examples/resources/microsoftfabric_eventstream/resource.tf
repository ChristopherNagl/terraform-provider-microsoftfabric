resource "microsoftfabric_eventstream" "example_eventstream" {
  workspace_id = microsoftfabric_workspace.example.id
  name         = "Eventstream_demo"
  description  = "An eventstream description."
}