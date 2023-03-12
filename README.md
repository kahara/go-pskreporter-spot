# go-pskreporter-spot

Go library for submitting spots to
[pskreporter.info](https://pskreporter.info/).

See PSK Reporter's
["for Developers"](https://pskreporter.info/pskdev.html)
section for an overview about the system.

**FIXME** link to a blog post

## Testing

**FIXME** There ~~are~~ will be unit tests, and these ~~can~~ will be run with:

```console
go test .
```

There's an "integration" test that attempts to verify that
[PSKReporter/rs-pskreporter-demo](https://github.com/PSKReporter/rs-pskreporter-demo)
can ingest the generated spots correctly. Get
[Compose v2](https://github.com/docker/compose)
and run:

```console
docker compose build \
    && docker compose up --exit-code-from integration-test
```
