# Scaleway Containers Deploy

This github action makes it easy to deploy on the containers product from the path of a registry scaleway.
It adds the functionality of using its own domain which must be on scaleway.

You can check the `action.yaml` file to see how it works.

# Exemple of use:

scw_access_key, scw_secret_key & scw_containers_namespace_id will always be necessary 

## simple deploy

| input name                  | value                                  |
| --------------------------- | -------------------------------------- |
| type                        | deploy                                 |
| scw_registry                | rg.fr-par.scw.cloud/test/images:latest |

## simple teardown

| input name                  | value                                  |
| --------------------------- | -------------------------------------- |
| type                        | teardown                               |
| scw_registry                | rg.fr-par.scw.cloud/test/images:latest |

## dns deploy

| input name                  | value                                  |
| --------------------------- | -------------------------------------- |
| type                        | deploy                                 |
| scw_registry                | rg.fr-par.scw.cloud/test/images:latest |
| scw_dns                     | containers.test.fr                     |
| scw_dns_prefix (optional)   | testing                                |


if not define scw_dns_prefix, the action will use the default value: "name of you created container"

This created containers will be based on the tag name of the registry.

## dns teardown

| input name                  | value                                  |
| --------------------------- | -------------------------------------- |
| type                        | teardown                               |
| scw_registry                | rg.fr-par.scw.cloud/test/images:latest |
| scw_dns                     | containers.test.fr                     |
| scw_dns_prefix (optional)   | testing                                |
