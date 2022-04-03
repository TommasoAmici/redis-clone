# Redis clone written in Go

I wrote this as a learning exercise. It's a Redis compliant server written in Go.
I plan on implementing all functions available in Redis v1.

You can test it with `redis-cli` or `telnet`, for example:

```sh
make build
make run

redis-cli -p 6380 set 1 234
# OK
redis-cli -p 6380 get 1
# "234"
```

## Implemented commands

This is the list of commands that were available in Redis v1.

- [ ] AUTH
- [ ] BGREWRITEAOF
- [ ] BGSAVE
- [ ] DBSIZE
- [ ] DEBUG
- [X] DECR
- [X] DECRBY
- [X] DEL
- [X] ECHO
- [X] EXISTS
- [ ] EXPIRE
- [X] FLUSHALL
- [X] FLUSHDB
- [X] GET
- [ ] GETSET
- [X] INCR
- [X] INCRBY
- [ ] INFO
- [ ] KEYS
- [ ] LASTSAVE
- [ ] LINDEX
- [ ] LLEN
- [ ] LPOP
- [ ] LPUSH
- [ ] LRANGE
- [ ] LREM
- [ ] LSET
- [ ] LTRIM
- [ ] MGET
- [ ] MONITOR
- [X] MOVE
- [X] PING
- [X] QUIT
- [X] RANDOMKEY
- [ ] RENAME
- [ ] RENAMENX
- [ ] RPOP
- [ ] RPUSH
- [ ] SADD
- [ ] SAVE
- [ ] SCARD
- [ ] SDIFF
- [ ] SDIFFSTORE
- [X] SELECT
- [X] SET
- [ ] SETNX
- [ ] SHUTDOWN
- [ ] SINTER
- [ ] SINTERSTORE
- [ ] SISMEMBER
- [ ] SLAVEOF
- [ ] SMEMBERS
- [ ] SMOVE
- [ ] SORT
- [ ] SPOP
- [ ] SRANDMEMBER
- [ ] SREM
- [ ] SUBSTR
- [ ] SUNION
- [ ] SUNIONSTORE
- [ ] SYNC
- [ ] TTL
- [ ] TYPE

## Benchmarks

I used [`hyperfine`](https://github.com/sharkdp/hyperfine) to figure out if my implementation
is in line with Redis. I haven't actually looked at their source code as I didn't want
to be influenced by it when coming up with my own implementation, as that would defeat
any learning objective I had.

**Benchmark 1:** `redis-cli -p 6380 set 2 abcdefghijklmnopqrstuvwxyz`

```text
Time (mean ± σ):       5.8 ms ±   0.4 ms    [User: 1.9 ms, System: 1.2 ms]
Range (min … max):     4.9 ms …   7.9 ms    473 runs
```

**Benchmark 2:** `redis-cli 2 abcdefghijklmnopqrstuvwxyz`

```text
Time (mean ± σ):       5.9 ms ±   0.5 ms    [User: 1.9 ms, System: 1.2 ms]
Range (min … max):     5.0 ms …  10.0 ms    462 runs
```

**Summary**
  `redis-cli -p 6380 set 2 abcdefghijklmnopqrstuvwxyz` ran
    1.02 ± 0.12 times faster than `redis-cli 2 abcdefghijklmnopqrstuvwxyz`
