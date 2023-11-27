package provider_test

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-kypo/internal/provider"
)

const (
	providerConfig = `
provider "kypo" {
  endpoint  = "https://images.crp.kypo.muni.cz"
  client_id = "bzhwmbxgyxALbAdMjYOgpolQzkiQHGwWRXxm"
}
`
	gitlabProviderConfig = `
provider "gitlab" {
  base_url = "https://gitlab.ics.muni.cz/api/v4"
}
`
	gitlabTestingDefinition = gitlabProviderConfig + `
variable "TAG_NAME" {}

resource "gitlab_project_tag" "terraform_testing_definition" {
  name    = var.TAG_NAME
  ref     = "master"
  project = "5211"
}

resource "kypo_sandbox_definition" "test" {
  url = "git@gitlab.ics.muni.cz:muni-kypo-crp/prototypes-and-examples/sandbox-definitions/terraform-provider-testing-definition.git"
  rev = gitlab_project_tag.terraform_testing_definition.name
}
`
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"kypo": providerserver.NewProtocol6WithError(provider.New("test")()),
}

var gitlabProvider = map[string]resource.ExternalProvider{
	"gitlab": {
		Source:            "gitlabhq/gitlab",
		VersionConstraint: "15.11.0",
	},
}

//func testAccPreCheck(t *testing.T) {
// You can add code here to run prior to any test case execution, for example assertions
// about the appropriate environment variables being set are common to see in a pre-check
// function.
//}
