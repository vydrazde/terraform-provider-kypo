resource "kypo_sandbox_definition" "example" {
  url = "git@gitlab.ics.muni.cz:muni-kypo-trainings/games/junior-hacker.git"
  rev = "master"
}

resource "kypo_sandbox_pool" "example" {
  definition = {
    id = kypo_sandbox_definition.example.id
  }
  max_size = 2
}
