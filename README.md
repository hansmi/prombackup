# Backup for Prometheus

[![Latest release](https://img.shields.io/github/v/release/hansmi/prombackup)][releases]
[![Release workflow](https://github.com/hansmi/prombackup/actions/workflows/release.yaml/badge.svg)](https://github.com/hansmi/prombackup/actions/workflows/release.yaml)
[![CI workflow](https://github.com/hansmi/prombackup/actions/workflows/ci.yaml/badge.svg)](https://github.com/hansmi/prombackup/actions/workflows/ci.yaml)
[![Go reference](https://pkg.go.dev/badge/github.com/hansmi/prombackup.svg)](https://pkg.go.dev/github.com/hansmi/prombackup)

A command line utility and daemon to create and download snapshots of
a [Prometheus](https://prometheus.io/) server time-series database as an
archive.

[Prometheus' HTTP API](https://prometheus.io/docs/prometheus/latest/querying/api/)
exposes an endpoint to create snapshots of its time-series database. The
snapshots are only available on the filesystem local to the server and can't be
retrieved easily. Prombackup combines taking a snapshot, sending its content as
a [tarball](https://en.wikipedia.org/wiki/Tar_%28computing%29) to a client and
cleaning up unused snapshots.

Prombackup includes a rudimentary web interface for interactive usage. The main
interface is the `prombackup` command line utility.

Users wanting to implement authentication and authorization must do so using
a reverse proxy in front of the Prombackup server.


## Usage

Launch the server:

```shell
prombackup-server \
  -listen_address :8080 \
  -prometheus_endpoint http://prometheus:9090 \
  -snapshot_dir /storage/snapshots
```

The snapshot directory must be shared with Prometheus and `prombackup-server`
must have read access. Pruning snapshots requires write access.

The most important flags can be configured via environment variables. See the
output of `prombackup-server -help` for additional information.

Create a new snapshot and download it to the current directory:

```shell
prombackup -server http://prombackupserver:8080 create
```

The server URL can be controlled via the `PROMBACKUP_ENDPOINT` environment
variable. Fetch a snapshot into a particular file using server-side
compression:

```shell
PROMBACKUP_ENDPOINT=http://prombackupserver:8080 \
prombackup create -format tgz -output /tmp/mybackup.tar.gz
```

Prometheus snapshots consist of
[hard links](https://en.wikipedia.org/wiki/Hard_link) to the time-series
database. They don't consume significant amounts of filesystem space on their
own except when referring to expired data. Pruning automatically removes
snapshots older than a certain amount of time:

```shell
PROMBACKUP_ENDPOINT=http://prombackupserver:8080 \
prombackup prune -keep_within 12h
```

The server can be configured to automatically prune in regular intervals using
its `-autoprune` flag.


## Installation

Pre-built binaries are provided for [all releases][releases]:

* Binary archives (`.tar.gz`)
* Debian/Ubuntu (`.deb`)
* RHEL/Fedora (`.rpm`)

With the source being available it's also possible to produce custom builds
directly using [Go](https://go.dev/) or [GoReleaser](https://goreleaser.com/).


## Implementation considerations

Downloading a snapshot archive does not require any disk space. The archive is
built on-the-fly. A downside to this is that errors occurring after sending the
HTTP header can't be reported to the client. The client will only see
a terminated connection/request. The command line utility makes use of the
`X-Prombackup-Download-Id` header sent along with the HTTP response to look up
the status after a download is finished. A SHA-256 checksum of the response
body is also verified. This means that a download via a browser or another HTTP
client requires separate verification. The web interface lists links to status
information on the most recent downloads and the status endpoint can also be
invoked directly with the ID from the aforementioned header.

For simplicity the implementation is stateful and assumes that there's at most
one Prombackup instance per Prometheus server instance. Status information on
finished downloads (e.g. an error or the content checksum) is only kept in
memory.

The snapshot directory must be shared between Prometheus and Prombackup at
a filesystem level (network filesystem would work too).

[releases]: https://github.com/hansmi/prombackup/releases/latest

<!-- vim: set sw=2 sts=2 et : -->
