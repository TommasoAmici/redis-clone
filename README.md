# Redis clone written in Go

I wrote this as a learning exercise. It's a Redis compliant server written in Go.

You can test it with `redis-cli`, for example:

```sh
make build
make run

redis-cli -p 6380 set 1 234
# OK
redis-cli -p 6380 get 1
# "234"
```

## Implemented commands

- GET
- SET
- DEL
- EXISTS
- PING
- QUIT

## Benchmarks

I used [`hyperfine`](https://github.com/sharkdp/hyperfine) to figure out if my implementation
is in line with the Redis. I haven't actually looked at their source code as I didn't
want to be influenced by it when coming up with my own implementation.

**Benchmark 1:** `redis-cli -p 6380 set 2 abcdefghijklmnopqrstuvwxyz`

```
Time (mean ± σ):       5.8 ms ±   0.4 ms    [User: 1.9 ms, System: 1.2 ms]
Range (min … max):     4.9 ms …   7.9 ms    473 runs
```

**Benchmark 2:** `redis-cli 2 abcdefghijklmnopqrstuvwxyz`

```
Time (mean ± σ):       5.9 ms ±   0.5 ms    [User: 1.9 ms, System: 1.2 ms]
Range (min … max):     5.0 ms …  10.0 ms    462 runs
```

**Summary**
  `redis-cli -p 6380 set 2 abcdefghijklmnopqrstuvwxyz` ran
    1.02 ± 0.12 times faster than `redis-cli 2 abcdefghijklmnopqrstuvwxyz`
