# go-pskreporter-spot

Go library for submitting spots to pskreporter.info

See PSK Reporter's
["for Developers"](https://pskreporter.info/pskdev.html)
section for an overview about the system.

## Testing

There are unit tests, and these can be run with:

```console
go test ./...
```

To manually check that spot packets get ingested and processed by PSK Reporter, get
[Compose v2](https://github.com/docker/compose)
and run:

```console
docker compose up
```

This generates a few fake spots and sends them towards PSK Reporter's testing endpoint at
`pskreporter.info:14739`, then exits. The results can (hopefully!) be seen on PSK Reporter's
[packet analysis](https://pskreporter.info/cgi-bin/psk-analysis.pl)
page.

The composition includes an instance of
[tcpdump(1)](https://www.tcpdump.org/)
which dumps all traffic going to `14739/udp`.
