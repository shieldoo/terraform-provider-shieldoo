package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServerDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccServerDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.shieldoo_server.test", "id", "mockup"),
				),
			},
		},
	})
}

const testAccServerDataSourceConfig = `
provider "shieldoo" {
	endpoint = "https://mockup"
	apikey = "mockup"
}
data "shieldoo_server" "test" {
  name = "example"
}
`
