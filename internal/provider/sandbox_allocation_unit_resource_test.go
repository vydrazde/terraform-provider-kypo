package provider_test

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

data "kypo_sandbox_request_output" "test-user" {
  id = kypo_sandbox_allocation_unit.test.allocation_request.id
}
data "kypo_sandbox_request_output" "test-networking" {
  id = kypo_sandbox_allocation_unit.test.allocation_request.id
  stage = "networking-ansible"
}
data "kypo_sandbox_request_output" "test-terraform" {
  id = kypo_sandbox_allocation_unit.test.allocation_request.id
  stage = "terraform"
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

					// Datasource sandbox request output
					resource.TestCheckResourceAttrPair("data.kypo_sandbox_request_output.test-user", "id",
						"kypo_sandbox_allocation_unit.test", "allocation_request.id"),
					resource.TestCheckResourceAttr("data.kypo_sandbox_request_output.test-user", "stage", "user-ansible"),
					resource.TestCheckResourceAttrSet("data.kypo_sandbox_request_output.test-user", "result"),

					resource.TestCheckResourceAttrPair("data.kypo_sandbox_request_output.test-networking", "id",
						"kypo_sandbox_allocation_unit.test", "allocation_request.id"),
					resource.TestCheckResourceAttr("data.kypo_sandbox_request_output.test-networking", "stage", "networking-ansible"),
					resource.TestCheckResourceAttrSet("data.kypo_sandbox_request_output.test-networking", "result"),

					resource.TestCheckResourceAttrPair("data.kypo_sandbox_request_output.test-terraform", "id",
						"kypo_sandbox_allocation_unit.test", "allocation_request.id"),
					resource.TestCheckResourceAttr("data.kypo_sandbox_request_output.test-terraform", "stage", "terraform"),
					resource.TestCheckResourceAttrSet("data.kypo_sandbox_request_output.test-terraform", "result"),
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
