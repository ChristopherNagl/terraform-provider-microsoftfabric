resource "microsoftfabric_semantic_model_user_assignment" "example_assignment" {
  group_id   = "5225cb24-9857-4bc7-b556-7065b1f2daa6"
  dataset_id = "70673702-b9ed-457f-b199-da897b11edaa"
  users = [

    {
      email          = "AlexW@3cg7y4.onmicrosoft.com"
      role           = "ReadReshare"
      principal_type = "User"
    }
  ]
}