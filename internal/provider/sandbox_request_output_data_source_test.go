package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSandboxRequestOutputDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders:        gitlabProvider,
		Steps: []resource.TestStep{
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
		},
	})
}
