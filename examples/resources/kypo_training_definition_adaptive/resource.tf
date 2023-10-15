resource "kypo_training_definition_adaptive" "example" {
  content = jsonencode(
    {
      title              = "test"
      description        = null
      state              = "UNRELEASED"
      show_stepper_bar   = true
      estimated_duration = 0
      outcomes           = []
      phases             = []
      prerequisites      = []
    }
  )
}
