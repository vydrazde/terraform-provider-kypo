---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kypo_sandbox_pool Resource - terraform-provider-kypo"
subcategory: ""
description: |-
  Sandbox pool
---

# kypo_sandbox_pool (Resource)

Sandbox pool



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `definition` (Attributes) The associated sandbox definition (see [below for nested schema](#nestedatt--definition))
- `max_size` (Number) Maximum number of sandboxes

### Read-Only

- `created_by` (Attributes) Creator of this sandbox pool (see [below for nested schema](#nestedatt--created_by))
- `hardware_usage` (Attributes) Current resource usage (see [below for nested schema](#nestedatt--hardware_usage))
- `id` (Number) Sandbox Pool Id
- `lock_id` (Number) Id of associated lock
- `rev` (String) Revision of the associated Git repository of the sandbox definition
- `rev_sha` (String) Revision hash of the Git repository of the sandbox definition
- `size` (Number) Current number of sandboxes

<a id="nestedatt--definition"></a>
### Nested Schema for `definition`

Required:

- `id` (Number) Id of associated sandbox definition

Read-Only:

- `created_by` (Attributes) Creator of this sandbox definition (see [below for nested schema](#nestedatt--definition--created_by))
- `name` (String) Name of the sandbox definition
- `rev` (String) Revision of the Git repository of the sandbox definition
- `url` (String) Url to the Git repository of the sandbox definition

<a id="nestedatt--definition--created_by"></a>
### Nested Schema for `definition.created_by`

Read-Only:

- `family_name` (String) TODO
- `full_name` (String) TODO
- `given_name` (String) TODO
- `id` (Number) Id of the user
- `mail` (String) TODO
- `sub` (String) TODO



<a id="nestedatt--created_by"></a>
### Nested Schema for `created_by`

Read-Only:

- `family_name` (String) TODO
- `full_name` (String) TODO
- `given_name` (String) TODO
- `id` (Number) Id of the user
- `mail` (String) TODO
- `sub` (String) TODO


<a id="nestedatt--hardware_usage"></a>
### Nested Schema for `hardware_usage`

Read-Only:

- `instances` (String) TODO
- `network` (String) TODO
- `port` (String) TODO
- `ram` (String) TODO
- `subnet` (String) TODO
- `vcpu` (String) Id of the user

