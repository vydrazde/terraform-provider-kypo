package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSandboxPoolResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "kypo_sandbox_definition" "test" {
  url = "git@gitlab.ics.muni.cz:muni-kypo-trainings/games/junior-hacker.git"
  rev = "master"
}
resource "kypo_sandbox_pool" "test" {
  definition = {
    id = kypo_sandbox_definition.test.id
  }
  max_size = 2
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kypo_sandbox_pool.test", "size", "0"),
					resource.TestCheckResourceAttr("kypo_sandbox_pool.test", "max_size", "2"),
					resource.TestCheckResourceAttr("kypo_sandbox_pool.test", "rev", "master"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "rev_sha"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.vcpu"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.ram"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.instances"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.network"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.subnet"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.port"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.sub"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.full_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.given_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.family_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.mail"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kypo_sandbox_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "kypo_sandbox_definition" "test" {
  url = "git@gitlab.ics.muni.cz:muni-kypo-trainings/games/junior-hacker.git"
  rev = "master"
}
resource "kypo_sandbox_pool" "test" {
  definition = {
    id = kypo_sandbox_definition.test.id
  }
  max_size = 10
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kypo_sandbox_pool.test", "size", "0"),
					resource.TestCheckResourceAttr("kypo_sandbox_pool.test", "max_size", "10"),
					resource.TestCheckResourceAttr("kypo_sandbox_pool.test", "rev", "master"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "rev_sha"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.vcpu"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.ram"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.instances"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.network"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.subnet"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "hardware_usage.port"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.sub"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.full_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.given_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.family_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_pool.test", "created_by.mail"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
