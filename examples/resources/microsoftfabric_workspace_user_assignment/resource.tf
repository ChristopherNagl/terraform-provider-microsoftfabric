resource "microsoftfabric_workspace_user_assignment" "example_assignment" {
  workspace_id = microsoftfabric_workspace.example.id # Replace with your workspace ID

  users = [
    {
      email          = "AdeleV@3cg7y4.onmicrosoft.com"
      role           = "Member"
      principal_type = "User"
    },
    {
      email          = "f4c6053c-5243-4690-9e1f-f1b5a7558202"
      role           = "Contributor"
      principal_type = "Group"
    }
  ]
}