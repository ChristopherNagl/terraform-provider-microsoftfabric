resource "microsoftfabric_lakehouse_table" "example" {
  workspace_id  = microsoftfabric_workspace.example.id
  lakehouse_id  = microsoftfabric_lakehouse.example_lakehouse2.id
  table_name    = "haha"
  relative_path = "Files/data/sales.csv"
  path_type     = "File"
  mode          = "Overwrite"
  recursive     = false
  format_options = {
    format    = "Csv"
    header    = true
    delimiter = ","
  }
}


resource "microsoftfabric_lakehouse_table" "example2" {
  workspace_id  = microsoftfabric_workspace.example.id
  lakehouse_id  = microsoftfabric_lakehouse.example_lakehouse2.id
  table_name    = "haha"
  relative_path = "${microsoftfabric_shortcut.example_shortcut.id}/sales.csv"
  path_type     = "File"
  mode          = "Overwrite"
  recursive     = false
  format_options = {
    format    = "Csv"
    header    = true
    delimiter = ","
  }
}
