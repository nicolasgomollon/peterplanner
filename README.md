# peterplanner

PeterPlanner is a course prerequisite tracker and scheduling tool for UCI students.

This is the core script that deals with generating a JSON output containing all of the relevant content for a user. This repo also contains the functions for scraping the necessary pages from [DegreeWorks](https://www.reg.uci.edu/access/student/degreeworks/?seg=U), [UCI General Catalogue](http://catalogue.uci.edu/), [Course Prerequisites](https://www.reg.uci.edu/cob/prrqcgi), and [WebSOC (Schedule of Classes)](https://www.reg.uci.edu/perl/WebSoc).


## See Also

- [nicolasgomollon/peterplanner.com](https://github.com/nicolasgomollon/peterplanner.com)

- [nicolasgomollon/peterplanner-download](https://github.com/nicolasgomollon/peterplanner-download)

- [nicolasgomollon/peterplanner-parse](https://github.com/nicolasgomollon/peterplanner-parse)


## System Requirements

The web app currently runs on a [LAMP stack](https://en.wikipedia.org/wiki/LAMP_(software_bundle)).


## Database Configuration

The application reads the `db` object in `/var/www/config/settings.json`.
For a MySQL container named `mysql` on the same Docker Compose network, use:

```json
{
  "db": {
    "kind": "mysql",
    "name": "app-peterplanner",
    "username": "YOUR_DATABASE_USER",
    "password": "YOUR_DATABASE_PASSWORD",
    "host": "mysql",
    "port": 3306
  }
}
```

Merge these fields into your existing configuration; keep any other settings.
`host` is optional and defaults to `127.0.0.1` when omitted or empty. `port` is
an optional JSON number and defaults to `3306` when omitted or zero. Supply
IPv6 hosts without brackets, for example `"host": "::1"`; the DSN formatter
adds the necessary address brackets. Existing files with only `kind`, `name`,
`username`, and `password` continue to use local TCP MySQL. The existing
`/var/www/...` paths still need to be available through your container mounts.

Install `github.com/go-sql-driver/mysql` in your build environment's GOPATH.
Use v1.4.1 or newer: v1.4.1 is the smallest release update with
`caching_sha2_password` support and its initial authentication fixes, and
provides the `mysql.NewConfig` API used here. Pin the selected version in your
Docker build; this repository does not vendor or manage that dependency.
The previous 2017 driver did not support MySQL 8.4's default authentication
and cannot compile the updated settings code.
No server authentication downgrade is needed. Full authentication over TCP
uses the driver's RSA password exchange when TLS is not configured; the server
must have working RSA keys. A server that requires TLS needs separate TLS
configuration; this change retains the existing connection options and does
not enable TLS. v1.4.1 is a minimal legacy compatibility version, not the latest
driver release.

These repositories retain their GOPATH dependency layout. Driver v1.4.1
requires Go 1.7 or newer and has no external dependencies; newer driver
versions may require newer Go toolchains and additional dependencies. With a
compatible driver and the existing dependencies available under GOPATH, test
and build the core without
running it:

```sh
GO111MODULE=off go test ./...
GO111MODULE=off go build -o /tmp/peterplanner .
```

Newer Go versions report an existing vet diagnostic for the nonconstant
`fmt.Printf` format in `main.go`. The database tests can be run independently
with `GO111MODULE=off go test ./helpers ./database`; to run the full test suite
despite that unrelated diagnostic, use `GO111MODULE=off go test -vet=off ./...`.

The settings tests use temporary files and parse DSNs without connecting to a
database. The downloader and parser do not use the database and require no
database configuration changes.


## License

PeterPlanner is released under the GNU AGPLv3.
