package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccFirewallResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("shieldoo_firewall.test", "id", "mockup"),
				),
			},
			// ImportState testing
			{
				ResourceName: "shieldoo_firewall.test",
				ImportState:  true,
			},
			// Update and Read testing
			{
				Config: testAccFirewallResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("shieldoo_firewall.test", "id", "mockup"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccFirewallResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
provider "shieldoo" {
	endpoint = "https://mockup"
	apikey = "mockup"
}
resource "shieldoo_firewall" "test" {
  name = %[1]q
}
`, configurableAttribute)
}
