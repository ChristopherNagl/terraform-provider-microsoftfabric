resource "microsoftfabric_domain" "example_domain" {
  display_name = "Finance"
  description  = "This domain is used for identifying financial data and reports.change"
}

resource "microsoftfabric_domain" "example_domain2" {
  display_name     = "Finance_Controlling"
  description      = "This domain is used for identifying financial data and reports.change"
  parent_domain_id = microsoftfabric_domain.example_domain.id
}
