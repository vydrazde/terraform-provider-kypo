package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSandboxAllocationUnitResource(t *testing.T) {
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
  max_size = 1
}

resource "kypo_sandbox_allocation_unit" "test" {
  pool_id = kypo_sandbox_pool.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kypo_sandbox_allocation_unit.test", "locked", "false"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "id"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_allocation_unit.test", "pool_id",
						"kypo_sandbox_pool.test", "id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "allocation_request.id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "allocation_request.allocation_unit_id"),
					resource.TestCheckResourceAttrPair("kypo_sandbox_allocation_unit.test", "allocation_request.allocation_unit_id",
						"kypo_sandbox_allocation_unit.test", "id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "allocation_request.created"),
					resource.TestCheckResourceAttr("kypo_sandbox_allocation_unit.test", "allocation_request.stages.#", "3"),
					resource.TestCheckResourceAttr("kypo_sandbox_allocation_unit.test", "allocation_request.stages.0", "FINISHED"),
					resource.TestCheckResourceAttr("kypo_sandbox_allocation_unit.test", "allocation_request.stages.1", "FINISHED"),
					resource.TestCheckResourceAttr("kypo_sandbox_allocation_unit.test", "allocation_request.stages.2", "FINISHED"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "created_by.id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "created_by.sub"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "created_by.full_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "created_by.given_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "created_by.family_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_allocation_unit.test", "created_by.mail"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kypo_sandbox_allocation_unit.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
