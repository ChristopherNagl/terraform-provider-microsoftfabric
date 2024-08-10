resource "microsoftfabric_workspace_git" "example_git" {
  workspace_id = microsoftfabric_workspace.example.id

  git_provider_details = {
    organization_name = "somename"
    project_name      = "ChrisFabric"
    git_provider_type = "AzureDevOps"
    repository_name   = "DevOps"
    branch_name       = "main"
    directory_name    = "/"
  }

  initialization_strategy = "PreferRemote"
}