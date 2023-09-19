package provider

//
//import (
//	"testing"
//
//	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
//)
//
//const ltdDefinition = `
//{
//  "title" : "test",
//  "description" : null,
//  "prerequisites" : [ ],
//  "outcomes" : [ ],
//  "state" : "UNRELEASED",
//  "show_stepper_bar" : true,
//  "levels" : [ ],
//  "estimated_duration" : 0,
//  "variant_sandboxes" : false
//}
//`
//
//func TestAccTrainingDefinitionResource(t *testing.T) {
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		ExternalProviders:        gitlabProvider,
//		Steps: []resource.TestStep{
//			// Create and Read testing
//			{
//				Config: providerConfig + `
//resource "kypo_training_definition" "test" {
//  content = <<EOL
//` + ltdDefinition + `EOL
//}`,
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("kypo_training_definition.test", "content", ltdDefinition),
//					resource.TestCheckResourceAttrSet("kypo_training_definition.test", "id"),
//				),
//			},
//			// ImportState testing
//			{
//				ResourceName:      "kypo_training_definition.test",
//				ImportState:       true,
//				ImportStateVerify: true,
//			},
//			// Delete testing automatically occurs in TestCase
//		},
//	})
//}
