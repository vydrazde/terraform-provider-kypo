resource "kypo_training_definition" "example" {
  content = <<-EOL
    {
      "title" : "test",
      "description" : null,
      "prerequisites" : [ ],
      "outcomes" : [ ],
      "state" : "UNRELEASED",
      "show_stepper_bar" : true,
      "levels" : [ ],
      "estimated_duration" : 0,
      "variant_sandboxes" : false
    }
    EOL
}
