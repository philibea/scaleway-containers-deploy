<p align="center">
    <img src="./logo.png" alt="Scaleway logo" />
</p>

<h1 align="center">The simplest way to deploy your app in Scaleway</h3>

<br />

# [Scaleway](https://www.scaleway.com/) GitHub Action

**Scaleway Containers Github Action** is a Github Action plugin allowing Scaleway users to integrate Containers within their CI nicely.

- Website: https://www.scaleway.com
- Console: https://console.scaleway.com
- Documentation: https://www.scaleway.com/en/docs

## ✅ Requirements

- A **Scaleway** account. [Sign up now](https://console.scaleway.com/register/) if you don't have any account yet.

## 📖 Installation

- Create an API key: [how to generate your API token?](https://www.scaleway.com/en/docs/console/my-project/how-to/generate-api-key)

- Setup a secret named `SCW_SECRET_KEY` & `SCW_ACCESS_KEY` within your repository `Secrets` section and set its value with output of the previous step.

- Setup a Registry. [Registry](https://www.scaleway.com/en/docs/faq/containerregistry)
  Actually only Scaleway Registry is available.

- Setup a Containers Namespace `SCW_CONTAINER_NAMESPACE_ID` within your repository `Secrets` section and set its value with your Scaleway account namespace.
  This Namespace is used inside the same Region of your registry.

You can can setup this namespace with our cli `scw containers namespace create` command.

- (optional) Setup a `SCW_DNS_ZONE` within your repository `Secrets` section and set its value with your Scaleway account DNS zone.

## 🔌 Usage

scw_access_key, scw_secret_key & scw_containers_namespace_id will always be necessary

### simple deploy

| input name         | value                                  |
| ------------------ | -------------------------------------- |
| type               | deploy (default value )                |
| scw_registry       | rg.fr-par.scw.cloud/test/images:latest |
| scw_container_port | 80 (default value )                    |

```bash
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    name: Deploy on Qovery
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Scaleway Container Deploy action
        id: deploy
        uses:  philibea/scaleway-containers-deploy@v1.0.5
        with:
          type: deploy
          scw_access_key:  ${{ secrets.ACCESS_KEY }}
          scw_secret_key: ${{ secrets.SECRET_KEY }}
          scw_containers_namespace_id: ${{ secrets.CONTAINERS_NAMESPACE_ID }}
          scw_registry: rg.fr-par.scw.cloud/test/testing:latest
          scw_container_port: "80"

```

### simple teardown

| input name   | value                                  |
| ------------ | -------------------------------------- |
| type         | teardown                               |
| scw_registry | rg.fr-par.scw.cloud/test/images:latest |

### dns deploy

| input name                | value                                  |
| ------------------------- | -------------------------------------- |
| type                      | deploy                                 |
| scw_registry              | rg.fr-par.scw.cloud/test/images:latest |
| scw_dns                   | containers.test.fr                     |
| scw_dns_prefix (optional) | testing                                |

if not define scw_dns_prefix, the action will use the default value: "name of you created container"

This created containers will be based on the tag name of the registry.

### dns teardown

| input name                | value                                  |
| ------------------------- | -------------------------------------- |
| type                      | teardown                               |
| scw_registry              | rg.fr-par.scw.cloud/test/images:latest |
| scw_dns                   | containers.test.fr                     |
| scw_dns_prefix (optional) | testing                                |
