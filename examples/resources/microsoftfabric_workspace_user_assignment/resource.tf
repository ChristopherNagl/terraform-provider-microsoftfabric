resource "microsoftfabric_workspace_user_assignment" "example_assignment" {
  workspace_id = microsoftfabric_workspace.example.id # Replace with your workspace ID

  users = [
    {
      email = "AdeleV@3cg7y4.onmicrosoft.com"
      role  = "Member"
    }
  ]
}