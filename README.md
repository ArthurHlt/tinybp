# TinyBp

Tinybp is a tiny bookmark provider for using inside reverse proxy or bookmark linker.

For now only [traefik](https://doc.traefik.io/traefik/) is implemented.
## Usage

1. donwload latest release and uncompress it
2. Create a `config.yml` with this content:
```yaml
listen: 0.0.0.0:8080
domain: my.domain.for.registration
bookmarks:
  - name: a-name # this will create Host a-name.my.domain.for.registration route on provider
    url: http://url.pointing.to.bookmark
    linker_config:
      entryPoints: [ "https" ]
```
3. set a [traefik configuration discovery](https://doc.traefik.io/traefik/providers/http/) pointing on http://you.tinybp.com/traefik

## Configuration options

- `[]` means optional (by default parameter is required)
- `<>` means type to use

### Root configuration in config.yml

```yaml
# Listen address for listening for http
[ listen: <string> | default = 0.0.0.0:8080 ]

# Domain for registering your bookmark
domain: <string>

log:
  # log level to use for server
  # you can chose: `trace`, `debug`, `info`, `warn`, `error`, `fatal` or `panic`
  [ level: <string> | default = info ]
  # Set to true to force not have color when seeing logs
  [ no_color: <bool> ]
  # et to true to see logs as json format
  [ in_json: <bool> ]


# list of entry (defined below)
bookmarks:
- <bookmark>
```

### Bookmark configuration

```yaml

# name identifier for creating url in bookmark
# it will become <name>.<domain>
name: <string>
# url for pointing on link to register
url: <string>
# By default bookmark are make a redirection to link
# If set to true, if bookmark linker support it, it will proxify url
[ proxify: <bool> ]
# If proxify is not set to true, this parameter is useless
# If true, request to url when proxying will not verify ssl certificate
[insecure_skip_verify: <bool>]
# linker config for passing it
# for now, only traefik use it
# you can set `entryPoints` for traefik and `enableTls` to true to enable resolve on traefik on tls also.
linker_config:
  <string>: <map|string|list>
```
