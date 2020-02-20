# `webpushy`

A 1 minute demo video is [viewable on Terminalizer.com][demo-video].

## CLI usage

```
webpushy usage:

webpushy send init --subscriber EMAIL
	generates keys, prints to terminal and writes to ~/.webpushy/keys.json

webpushy send --endpoint URL [--payload PAYL] [--ttl SECS] [--public KEY] [--private KEY] [--subscriber EMAIL]
	sends. falls back to ~/.webpushy/keys.json if keys are unspecified. ttl defaults
	to 0. payloads are one-per-line on stdin if unspecified.

webpushy recv init --name NAME --public KEY
	receives an endpoint ID and URL from push service. prints to terminal and writes to
	~/.webpushy/name.json.

webpush recv --name NAME [--limit COUNT] [--timeout SECS]
	connects to push service and streams messages to stdout. exits after COUNT
	number of messages or SECS total time spent running. otherwise runs forever
```

## Installation

* Mac: `brew install glassechidna/taps/webpushy`
* Windows: `scoop bucket add glassechidna https://github.com/glassechidna/scoop-bucket.git; scoop install webpushy`
* Otherwise get the latest build from the [Releases][releases] tab.

[demo-video]: https://terminalizer.com/view/2afec1ab3151
[releases]: https://github.com/glassechidna/webpushy/releases
