resource "microsoftfabric_pipeline" "example_pipeline" {
  display_name = "example pipeline"
  description  = "example pipeline"

  workspaces = [
    {
      workspace_id = microsoftfabric_workspace.example.id
      stage_order  = 1
    }
  ]
}