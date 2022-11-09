# Usage

## Available commands

### Get list of users

:::note
This command is only available for console super administrator
:::

```bash
$ miactl get users
```

### Create new user

:::note
This command is only available for console super administrator
:::

The parameters `fullname`, `firstname`, `lastname` are optional

```bash
$ miactl add user "john.smith@example.com" --fullname="John Smith" --firstname="John" --lastname="Smith"
```

### Delete existing user

:::note
This command is only available for console super administrator
:::

```bash
$ miactl delete user "john.smith@example.com"
```
