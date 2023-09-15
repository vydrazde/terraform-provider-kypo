package provider

//
//import (
//	"testing"
//
//	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
//)
//
//const atdDefinition = `
//{
//  "title" : "test",
//  "description" : null,
//  "prerequisites" : [ ],
//  "outcomes" : [ ],
//  "state" : "UNRELEASED",
//  "show_stepper_bar" : true,
//  "phases" : [ ],
//  "estimated_duration" : 0
//}
//`
//
//func TestAccTrainingDefinitionAdaptiveResource(t *testing.T) {
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		ExternalProviders:        gitlabProvider,
//		Steps: []resource.TestStep{
//			// Create and Read testing
//			{
//				Config: providerConfig + `
//resource "kypo_training_definition_adaptive" "test" {
//  content = <<EOL
//` + atdDefinition + `EOL
//}`,
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("kypo_training_definition_adaptive.test", "content", atdDefinition),
//					resource.TestCheckResourceAttrSet("kypo_training_definition_adaptive.test", "id"),
//				),
//			},
//			// ImportState testing
//			{
//				ResourceName:      "kypo_training_definition_adaptive.test",
//				ImportState:       true,
//				ImportStateVerify: true,
//			},
//			// Delete testing automatically occurs in TestCase
//		},
//	})
//}
