# Examples

## Create and use a Context

You can have one or more context locally for interacting with one or more installation fo Mia-Platform Console. Below
you can find some examples on how to create multiple contexts and then selecting one of them.

Create a Context for a company on the cloud instance:

```sh
miactl context set paas-company --company-id <your-company-id>
```

Create a Context for specific project in a company on the cloud instance:

```sh
miactl context set paas-project --company-id <your-company-id> --project-id <your-project-id>
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

## Deploy Project
