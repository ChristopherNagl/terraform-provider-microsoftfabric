resource "microsoftfabric_spark_pool" "example_spark_pool" {
  workspace_id = microsoftfabric_workspace.example.id
  name         = "example_spark_pool"
  node_family  = "MemoryOptimized"
  node_size    = "Small"

  auto_scale = {
    enabled        = true
    min_node_count = 1
    max_node_count = 3
  }

  dynamic_executor_allocation = {
    enabled       = true
    min_executors = 1
    max_executors = 2
  }
}