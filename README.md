# Terraform KYPO Provider

This [KYPO Provider]() allows [Terraform](https://www.terraform.io/) to manage [KYPO CRP](https://crp.kypo.muni.cz/) resources.

**This Terraform provider is in early development and not intended for use yet. Documentation is missing and breaking changes will occur.**


## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.18

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Install dependencies using the Go `mod tidy` command:
```shell
go mod tidy
```

4. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

1. Find the `GOBIN` path where Go installs your binaries. Your path may vary depending on how your Go environment variables are configured.
```shell
go env GOBIN
```
If the `GOBIN` go environment variable is not set, use the default path, `/home/<Username>/go/bin`.

2. Create a new file called `.terraformrc` in the root directory (`~`), then add the `dev_overrides` block below. 
Change the `<PATH>` to the value returned from the `echo $GOBIN` command above.
```
provider_installation {

  dev_overrides {
      "registry.terraform.io/vydrazde/kypo" = "<PATH>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

3. Set up your provider configuration. You can either copy the `examples/provider/provider.tf` file to a folder of one of the 
other `examples` and modify it, or fill the environment variables `KYPO_ENDPOINT`, `KYPO_CLIENT_ID`, `KYPO_USERNAME` and `KYPO_PASSWORD`.

4. Now you can use one of the examples in the `examples` directory and run
```shell
terraform plan
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up-to-date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
