resource "microsoftfabric_shortcut" "example_shortcut" {
  workspace_id = microsoftfabric_workspace.example.id
  item_id      = microsoftfabric_lakehouse.example_lakehouse.id # Replace with your actual item ID
  path         = "Files"                                        # Path where the shortcut will be created
  name         = "demo4"                                        # Name of the shortcut

  target = {
    adls_gen2 = {
      location      = "https://blobstoragedemoterragen.dfs.core.windows.net" # Replace accordingly
      subpath       = "/demo"                                                # Specify subpath if necessary
      connection_id = "9c08dcab-4d97-448d-8f42-7b0ba25bdc7a"                 # Replace with your connection ID
    }
  }
}