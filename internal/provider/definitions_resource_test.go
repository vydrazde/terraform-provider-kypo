package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccOrderResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "kypo_definitions" "test" {
  url = "git@gitlab.ics.muni.cz:muni-kypo-trainings/games/junior-hacker.git"
  rev = "master"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kypo_definitions.test", "url", "git@gitlab.ics.muni.cz:muni-kypo-trainings/games/junior-hacker.git"),
					resource.TestCheckResourceAttr("kypo_definitions.test", "rev", "master"),
					resource.TestCheckResourceAttr("kypo_definitions.test", "name", "junior-hacker-sandbox"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "id"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.id"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.sub"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.full_name"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.given_name"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.family_name"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.mail"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kypo_definitions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "kypo_definitions" "test" {
  url = "git@gitlab.ics.muni.cz:muni-kypo-crp/prototypes-and-examples/sandbox-definitions/kypo-crp-demo-training.git"
  rev = "master"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kypo_definitions.test", "url", "git@gitlab.ics.muni.cz:muni-kypo-crp/prototypes-and-examples/sandbox-definitions/kypo-crp-demo-training.git"),
					resource.TestCheckResourceAttr("kypo_definitions.test", "rev", "master"),
					resource.TestCheckResourceAttr("kypo_definitions.test", "name", "kypo-crp-demo-training"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "id"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.id"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.sub"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.full_name"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.given_name"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.family_name"),
					resource.TestCheckResourceAttrSet("kypo_definitions.test", "created_by.mail"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
