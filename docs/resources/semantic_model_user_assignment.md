---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "microsoftfabric_semantic_model_user_assignment Resource - microsoftfabric"
subcategory: ""
description: |-
  
---

# microsoftfabric_semantic_model_user_assignment (Resource)



## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `semantic_model_id` (String)
- `users` (Attributes List) (see [below for nested schema](#nestedatt--users))
- `workspace_id` (String)

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Required:

- `email` (String)
- `principal_type` (String) The principal type (App, Group, None, User)
- `role` (String)
