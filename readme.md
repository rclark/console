# console

Open the AWS console using credentials from your `~/.aws/config` file.

```
❯ console --help
Log in to the AWS console

Usage:
  console [flags]

Flags:
  -h, --help             help for console
  -p, --profile string   profile to use (default "default")
```

This will only work for profiles that are configured to perform `iam:AssumeRole` API calls to retreive AWS credentials. This includes SSO-based profiles.

## Install it

[Download the most recent release](https://github.com/rclark/console/releases) for your system's OS & architecture, unpack it, and place the `console` (or `console.exe`) file on your `$PATH`.
