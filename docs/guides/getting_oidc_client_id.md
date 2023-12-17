---
page_title: "How to get client_id of KYPO CRP"
---
[client_id](https://registry.terraform.io/providers/vydrazde/kypo/latest/docs#client_id) is one of the parameters for using the Terraform KYPO provider. As of KYPO version `23.12`, the default value `KYPO-Client` should work, for older KYPO instances or cases where the `client_id` has been changed follow this guide.

To get the `client_id` value for your KYPO CRP instance, visit the homepage of the KYPO instance in an anonymous window and open browser developer tools.

![image](https://github.com/vydrazde/terraform-provider-kypo/assets/80331839/60c9f152-e1c7-49e9-a386-80634b1f633a)

Click the `Login with local issuer` or `Login with local Keycloak` button and see one of the first network requests, where you will see the `client_id` among request headers.

![screenshot](https://github.com/vydrazde/terraform-provider-kypo/assets/80331839/a6a015d4-1e25-4aaa-895f-e265a171732f)
