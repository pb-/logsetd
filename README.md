# logsetd

logsetd implements a daemon for the logset protocol. Use [logset](https://github.com/pb-/logset) to synchronize two daemon instances.

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


### Synchronization process

Best explained [in code](https://github.com/pb-/logset/blob/ec6ca9a56844546d19d9af19968bb70fbc4a400c/logset/sync.py#L50) (less than 10 lines!) or in this ASCII-art graphic:

```
local                          sync                           remote

      < offsets()              (1)
      local_off >
                               (2)          pull(local_off) >
                                    < remote_off, remote_data
      < push(remote_data)      (3)

      < pull(remote_off)       (4)
      local_off', local_data >
                               (5)         push(local_data) >
```

Note that

 * the roles of `local` and `remote` are completely arbitrary. However, since there are fewer round trips on the right-hand side, it makes sense to have the remote peer on the right.
 * steps (3) and (5) will only happen if there is data to transfer, reducing the remote side to one round trip in the no-update case.
