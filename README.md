# datadog

CLI for Datadog API

## Installation

```bash
$ go get github.com/vsco/datadog
```

Ensure `$GOPATH/bin` is in your `$PATH` in order to use the binary.

## Authorization

In order to successfully use this tool, you will have to provide it a
datadog API key and application key. [Visit your DatadogHQ account settings](https://app.datadoghq.com/account/settings#api)
to generate a new application key.

Once you have these values, you have two options for configuration:
environment variables or a configuration file.

For **environment variables**, assign your API key to `DATADOG_API_KEY` and
your application key to `DATADOG_APP_KEY`.

For **configuration files**, put your API key and app key in a file like:

```json
{
  "api_key": "YOUR_API_KEY",
  "app_key": "YOUR_APP_KEY"
}
```

The default location for this file is `~/.datadogrc`, but you may put it
anywhere and tell the CLI about it via a flag: `-config=path/to/your/file`.

## Usage

```bash
$ datadog TYPE NAME VALUE
```

This CLI accepts with two types: `increment` and `gauge`. They send
a `counter` or `gauge` metric to Datadog, respectively.

The name of the metric is identical to the statsd metric name, e.g.
`vsco.my_metric`.

The value must be a float.

### Increment

```bash
$ datadog increment vsco.my_metric 100
```

`increment` can be shortened to `incr` or `i`, and one may use `counter` or
`c` as an alias.

### Gauge

```bash
$ datadog gauge vsco.my_metric 100
```

`gauge` can be shortened to `g`.

## License / Credit

This code is licensed under the MIT License, with credit to Visual Supply
Co (VSCO). See [LICENSE](LICENSE) for more.
