---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "microsoftfabric_eventstream Resource - microsoftfabric"
subcategory: ""
description: |-
  
---

# microsoftfabric_eventstream (Resource)



## Example Usage

```terraform
resource "microsoftfabric_eventstream" "example_eventstream" {
  workspace_id = microsoftfabric_workspace.example.id
  name         = "Eventstream_demo"
  description  = "An eventstream description."
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `description` (String)
- `name` (String)
- `workspace_id` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `last_updated` (String)
