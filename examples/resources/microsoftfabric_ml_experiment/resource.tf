resource "microsoftfabric_ml_experiment" "example_ml_experiment" {
  workspace_id = microsoftfabric_workspace.example.id
  display_name = "ml_experiment_demo"
  description  = "An example_ml_experiment description"
}