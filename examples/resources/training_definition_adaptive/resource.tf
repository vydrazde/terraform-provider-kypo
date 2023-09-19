resource "kypo_training_definition_adaptive" "example" {
  content = <<-EOL
    {
      "title" : "test",
      "description" : null,
      "prerequisites" : [ ],
      "outcomes" : [ ],
      "state" : "UNRELEASED",
      "show_stepper_bar" : true,
      "phases" : [ ],
      "estimated_duration" : 0
    }
    EOL
}
