# logsetd

logsetd implements a daemon for the logset protocol.

```shell
make develop  # build & run server
make test     # run tests
make build    # build binary
```


## logset Protocol

logset is a protocol to synchronize a set of append-only commit logs. Logs contain arbitrary byte streams. This section describes logset-over-http(s).

Endpoint              | Request body message | Response body message
----------------------|----------------------|----------------------
`GET /:repo/offsets`  |                      | `offsets`
`POST /:repo/pull`    | `offsets`            | `pull-response`
`POST /:repo/push`    | `push`               |


### Message grammar

The protocol mixes binary and text in the general case; text should be decoded as ASCII.

```
offsets := (name ' ' integer '\n')* '\n'
pull-response := offsets slice*
push := slice*
slice := slice-info slice-body
slice-info := name ' ' integer ' ' integer '\n'
slice-body := byte+
name := [0-9a-zA-Z]+
integer := [0-9]*
byte is an arbitrary byte
```
