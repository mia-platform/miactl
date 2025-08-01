# Examples

## Create and use a Context

You can have one or more context locally for interacting with one or more installation fo Mia-Platform Console. Below
you can find some examples on how to create multiple contexts and then selecting one of them.

Create a Context for a Company on the cloud instance:

```sh
miactl context set paas-company --company-id <your-company-id> --endpoint https://console.cloud.mia-platform.eu
```

Create a Context for specific Project in a Company on the cloud instance:

```sh
miactl context set paas-project --company-id <your-company-id> --project-id <your-project-id> --endpoint https://console.cloud.mia-platform.eu
```

Create a Context for connecting on a self hosted instance:

```sh
miactl context set example-console --endpoint https://example.com
```

Create a Context for connecting on a self hosted instance exposed via a self signed certificate:

```sh
miactl context set example-private --endpoint https://console.private --ca-cert /path/to/custom/private/ca.crt
```

Use the context named `paas-project`:

```sh
miactl context use paas-project
```

## List Projects

The `list project` command will list all Projects that the current user can access in the selected Company:

```sh
miactl project list
```

Or you can set a different Company via flag:

```sh
miactl project list --company-id <your-company-id>
```

## Deploy Project

The deploy command allows you to trigger a new deploy pipeline for the current Environment in the Project. The only
argument needed is the Environment ID that you want to deploy:

```sh
miactl deploy development --revision main
```

Additionally, if your context doesn’t contain the Project ID, you can select it via a flag:

```sh
miactl deploy development --project-id <your-project-id> --revision main
```

You can customize the way your Project is deployed:

```sh
miactl deploy development --no-semver --revision tags/v1.0.0
```

## Deploy Project using a service account from a CD pipeline

The commands are the same used above in the [Deploy Project](#deploy-project) section, but you need to use a
_Service Account_ for that.  
If you don't know how to create a _Service Account_, read the [dedicated documentation](https://docs.mia-platform.eu/docs/development_suite/identity-and-access-management/manage-service-accounts).

The _Service Account_ can be created with [two different authentication methods](https://docs.mia-platform.eu/docs/development_suite/identity-and-access-management/manage-service-accounts#adding-a-service-account):

* _Client Secret Basic_: the service account authenticates by presenting its `client_id` and `client_secret`;
* _Private Key JWT_: the service account authenticates by signing a `JWT` (JSON Web Token) using its private key.

After creating the _Service Account_, the first step to setup the `miactl` is **create an auth context**.
With an _auth context_ you can choose how to be authenticated with the Mia-Platform APIs in all your different contexts
you create with the `miactl`.

Based on the authentication method of your _Service Account_, you can create the auth context with the following command:

* _Client Secret Basic_:

  ```sh
  miactl context auth <miactl-auth-name> --client-id <sa-client-id> --client-secret <sa-client-secret>
  ```

* _Private Key JWT_:

  ```sh
  miactl context auth <miactl-auth-name> --jwt-json <path-to-json-containing-the-json-config-of-a-jwt-service-account>
  ```

Now you can create the context you want use the `miactl` to.

:::warning
Remember to specify the auth context to be used with the `---auth-name` flag, otherwise the `miactl` will try to perform
a user authentication through the default browser.
:::

```sh
miactl context set <my-context-name> --endpoint https://console.private --company-id <my-company-id> --project-id <my-project-id> --auth-name <miactl-auth-name>
```

After that, just set the context as the used one:

```sh
miactl context use <my-context-name>
```

and deploy the pipeline:

```sh
miactl deploy development --no-semver --revision main
```

Finally, you can group the commands above and run them inside a pipeline, e.g. a GitLab pipeline:

```yaml
# Insert that after your pipeline stages
delivery:
    stage: deploy
    image: ghcr.io/mia-platform/miactl:v0.20.0

    script:
      - miactl version
      - miactl context auth deployer-sa --client-id sa-client-id --client-secret sa-super-secret
      - miactl context set my-private-console --endpoint https://console.private --company-id id-of-my-company --project-id id-of-my-project --auth-name deployer-sa
      - miactl use my-private-console
      - miactl deploy DEV --no-semver --deploy-type smart_deploy --revision main
```
