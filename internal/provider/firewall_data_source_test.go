package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccFirewallDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.shieldoo_firewall.test", "id", "mockup"),
				),
			},
		},
	})
}

const testAccFirewallDataSourceConfig = `
provider "shieldoo" {
	endpoint = "https://mockup"
	apikey = "mockup"
}
data "shieldoo_firewall" "test" {
  name = "example"
}
`
