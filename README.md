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

- Setup a [Registry](https://www.scaleway.com/en/docs/faq/containerregistry)
  Actually only Scaleway Registry is available.

- Setup a Containers Namespace `SCW_CONTAINER_NAMESPACE_ID` within your repository `Secrets` section and set its value with your Scaleway account namespace.
  This Namespace is used inside the same Region of your registry.

You can can setup this namespace with our cli `scw containers namespace create` command.

- (optional) Setup a `SCW_DNS_ZONE` within your repository `Secrets` section and set its value with your Scaleway account DNS zone.
How To add [Custom Domains](https://www.scaleway.com/en/docs/compute/containers/how-to/add-a-custom-domain-to-a-container/).
In this automation process, we will use the DNS zone of your Scaleway account. Each zone will be based on the container name created and based on the tag of your Image.
Your path registry is `rg.fr-par.scw.cloud/test/images:latest`, your container name tag will be `latest` and your DNS zone will be `latest.${SCW_DNS_ZONE}`.

## 🔌 Usage

`scw_access_key`, `scw_secret_key` & `scw_containers_namespace_id` will always be necessary

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
    name: Deploy on Scaleway Containers
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

```

### simple teardown

| input name   | value                                  |
| ------------ | -------------------------------------- |
| type         | teardown                               |
| scw_registry | rg.fr-par.scw.cloud/test/images:latest |

```bash
on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    name: Teardown Containers
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Scaleway Container Teardown action
        id: teardown
        uses:  philibea/scaleway-containers-deploy@v1.0.5
        with:
          type: teardown
          scw_access_key:  ${{ secrets.ACCESS_KEY }}
          scw_secret_key: ${{ secrets.SECRET_KEY }}
          scw_containers_namespace_id: ${{ secrets.CONTAINERS_NAMESPACE_ID }}
          scw_registry: rg.fr-par.scw.cloud/test/testing:latest

```

### dns deploy

| input name                | value                                  |
| ------------------------- | -------------------------------------- |
| type                      | deploy                                 |
| scw_registry              | rg.fr-par.scw.cloud/test/images:latest |
| scw_dns                   | containers.test.fr                     |

Actually, prefix of your dns will use the default value: "name of you created container"
This created containers will be based on the tag name of the registry.

```bash
on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    name: Deploy on Scaleway Containers
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
          scw_dns: containers.test.fr
```

### dns teardown

| input name                | value                                  |
| ------------------------- | -------------------------------------- |
| type                      | teardown                               |
| scw_registry              | rg.fr-par.scw.cloud/test/images:latest |
| scw_dns                   | containers.test.fr                     |
| scw_dns_prefix (optional) | testing                                |


```bash
on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    name: Teardown Containers
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Scaleway Container Teardown action
        id: teardown
        uses:  philibea/scaleway-containers-deploy@v1.0.5
        with:
          type: teardown
          scw_access_key:  ${{ secrets.ACCESS_KEY }}
          scw_secret_key: ${{ secrets.SECRET_KEY }}
          scw_containers_namespace_id: ${{ secrets.CONTAINERS_NAMESPACE_ID }}
          scw_registry: rg.fr-par.scw.cloud/test/testing:latest
          scw_dns: containers.test.fr
```


## 🐳 Docker

If you want to use this flow outside of Github Actions, you can use the Docker Image.

```bash
docker run -it --rm \
  -e INPUT_SCW_ACCESS_KEY=${SCW_ACCESS_KEY} \
  -e INPUT_SCW_SECRET_KEY=${SCW_SECRET_KEY} \
  -e INPUT_SCW_CONTAINERS_NAMESPACE_ID=${SCW_CONTAINERS_NAMESPACE_ID} \
  -e INPUT_SCW_REGISTRY=rg.fr-par.scw.cloud/test/testing:latest \
  -e INPUT_SCW_DNS=containers.test.fr \
  -e INPUT_TYPE=deploy \
  phiphi/scaleway-containers-deploy:latest
  ```

