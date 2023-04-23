package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSandboxPoolResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders:        gitlabProvider,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + gitlabTestingDefinition + `
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
					resource.TestCheckResourceAttr("kypo_sandbox_pool.test", "rev", os.Getenv("TF_VAR_TAG_NAME")),
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
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.id",
						"kypo_sandbox_pool.test", "definition.created_by.id"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.sub",
						"kypo_sandbox_pool.test", "definition.created_by.sub"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.full_name",
						"kypo_sandbox_pool.test", "definition.created_by.full_name"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.given_name",
						"kypo_sandbox_pool.test", "definition.created_by.given_name"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.family_name",
						"kypo_sandbox_pool.test", "definition.created_by.family_name"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.mail",
						"kypo_sandbox_pool.test", "definition.created_by.mail"),
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
				Config: providerConfig + gitlabTestingDefinition + `
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
					resource.TestCheckResourceAttr("kypo_sandbox_pool.test", "rev", os.Getenv("TF_VAR_TAG_NAME")),
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
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.id",
						"kypo_sandbox_pool.test", "definition.created_by.id"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.sub",
						"kypo_sandbox_pool.test", "definition.created_by.sub"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.full_name",
						"kypo_sandbox_pool.test", "definition.created_by.full_name"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.given_name",
						"kypo_sandbox_pool.test", "definition.created_by.given_name"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.family_name",
						"kypo_sandbox_pool.test", "definition.created_by.family_name"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_pool.test", "created_by.mail",
						"kypo_sandbox_pool.test", "definition.created_by.mail"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
