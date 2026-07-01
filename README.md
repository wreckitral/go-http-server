# go-http-server

An HTTP/1.1 server built on raw TCP in Go.

No `net/http`. Parses the request line, headers, and body manually
from a byte buffer over a `net.Conn`.

## What it handles

- `GET /` — 200 OK
- `GET /echo/<text>` — echoes text back, with gzip if client accepts it
- `GET /user-agent` — returns the client's User-Agent header
- `GET /files/<name>` — serves a file from a directory passed as CLI arg
- `POST /files/<name>` — writes request body to disk as a file
- anything else — 404

## Running it

```bash
make run DIR=./files
# or
go run . --directory ./files
```

## How the parsing works

Each connection reads raw bytes into a buffer, splits on `\r\n`,
then pulls the method and path from the request line. Headers are
scanned by string prefix. The body is whatever's left after the blank line separator.

For gzip: if the `Accept-Encoding` header contains `gzip`, the
echo endpoint compresses the response body with `compress/gzip`
and sets `Content-Encoding: gzip` accordingly.

## What I learned

HTTP is just text over TCP with a specific structure. Once you've
parsed it by hand you understand exactly what `net/http` is doing
for you and more importantly, what can go wrong when a request
comes in malformed or larger than your buffer.

## Stack

Go 1.22 · stdlib only · no external dependencies
