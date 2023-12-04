# Terraform KYPO Provider

Terraform provider for KYPO allows [Terraform](https://www.terraform.io/) to manage [KYPO CRP](https://crp.kypo.muni.cz/) resources.

See documentation at the [Terraform Registry](https://registry.terraform.io/providers/vydrazde/kypo/latest/docs).

## Example Usage

```terraform
terraform {
  required_providers {
    kypo = {
      source = "vydrazde/kypo"
      version = "0.3.1"
    }
  }
}

provider "kypo" {
  endpoint  = "https://your.kypo.ex" # Or use KYPO_ENDPOINT env var
  client_id = "***"                  # Or use KYPO_CLIENT_ID env var
  username  = "user"                 # Or use KYPO_USERNAME env var
  password  = "***"                  # Or use KYPO_PASSWORD env var
}

resource "kypo_sandbox_definition" "example" {
  url = "git@gitlab.ics.muni.cz:muni-kypo-trainings/games/junior-hacker.git"
  rev = "master"
}

resource "kypo_sandbox_pool" "example" {
  definition = {
    id = kypo_sandbox_definition.example.id
  }
  max_size = 1
}

resource "kypo_sandbox_allocation_unit" "example" {
  pool_id = kypo_sandbox_pool.example.id
}
```

# 
