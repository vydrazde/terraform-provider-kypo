resource "kypo_training_definition" "example" {
  content = jsonencode(
    {
      title              = "test"
      description        = null
      state              = "UNRELEASED"
      show_stepper_bar   = true
      variant_sandboxes  = false
      estimated_duration = 0
      levels             = []
      outcomes           = []
      prerequisites      = []
    }
  )
}
