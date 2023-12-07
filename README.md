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

# License
Parts of the Source Code, namely [GNUmakefile](./GNUmakefile), [.golangci.yml](./.golangci.yml), [.goreleaser.yml](./.goreleaser.yml),
[tools/tools.go](./tools/tools.go),
[internal/validators/timeduration.go](./internal/validators/timeduration.go) and [internal/validators/timeduration_test.go](./internal/validators/timeduration_test.go),
is subject to the terms of the Mozilla Public License, v. 2.0. You can obtain the license at https://mozilla.org/MPL/2.0/. 
The Source Code was used from [HashiCorp terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) and [HashiCorp terraform-plugin-framework-timeouts](https://github.com/hashicorp/terraform-plugin-framework-timeouts) repositories.

The rest of the Source Code is subject to the [MIT License](./LICENSE).

