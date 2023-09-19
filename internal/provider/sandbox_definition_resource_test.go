package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const gitlabTestingDefinitionTag = gitlabProviderConfig + `
variable "TAG_NAME" {}

resource "gitlab_project_tag" "terraform_testing_definition" {
  count   = 2

  name    = "${var.TAG_NAME}-${count.index}"
  ref     = "master"
  project = "5211"
}
`

func TestAccSandboxDefinitionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders:        gitlabProvider,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + gitlabTestingDefinitionTag + `
resource "kypo_sandbox_definition" "test" {
  url = "git@gitlab.ics.muni.cz:muni-kypo-crp/prototypes-and-examples/sandbox-definitions/terraform-provider-testing-definition.git"
  rev = gitlab_project_tag.terraform_testing_definition[0].name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kypo_sandbox_definition.test", "url", "git@gitlab.ics.muni.cz:muni-kypo-crp/prototypes-and-examples/sandbox-definitions/terraform-provider-testing-definition.git"),
					resource.TestCheckResourceAttr("kypo_sandbox_definition.test", "rev", os.Getenv("TF_VAR_TAG_NAME")+"-0"),
					resource.TestCheckResourceAttr("kypo_sandbox_definition.test", "name", "terraform-testing-definition"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.sub"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.full_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.given_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.family_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.mail"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kypo_sandbox_definition.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + gitlabTestingDefinitionTag + `
resource "kypo_sandbox_definition" "test" {
  url = "git@gitlab.ics.muni.cz:muni-kypo-crp/prototypes-and-examples/sandbox-definitions/terraform-provider-testing-definition.git"
  rev = gitlab_project_tag.terraform_testing_definition[1].name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kypo_sandbox_definition.test", "url", "git@gitlab.ics.muni.cz:muni-kypo-crp/prototypes-and-examples/sandbox-definitions/terraform-provider-testing-definition.git"),
					resource.TestCheckResourceAttr("kypo_sandbox_definition.test", "rev", os.Getenv("TF_VAR_TAG_NAME")+"-1"),
					resource.TestCheckResourceAttr("kypo_sandbox_definition.test", "name", "terraform-testing-definition"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.id"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.sub"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.full_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.given_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.family_name"),
					resource.TestCheckResourceAttrSet("kypo_sandbox_definition.test", "created_by.mail"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
