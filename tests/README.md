# How to write goc e2e test cases

Current goc e2e test is based on the [bats-core](https://github.com/bats-core/bats-core) framework, you should read its document first.

## Local dev requirements

* [bats-core](https://github.com/bats-core/bats-core)
* install `goc` to `PATH`
* build `goc` with `goc`, generate the binary `gocc`, install `gocc` to `PATH`

## Test goc

First of all, you should start a `goc server` in `setup_file` function from the backend, 

```
setup_file() {
    goc server 3>&- &
    GOC_PID=$!
    sleep 2
    goc init
}
```

According to [this](https://github.com/bats-core/bats-core/blob/master/README.md#file-descriptor-3-read-this-if-bats-hangs), you should turn off the file descriptor 3 for the long-running backend job. Then you can write any `goc` subcommand. Remember to kill the `$GOC_PID` in the `teardown_file` function.

## Test covered goc - gocc

We also need to test with the covered gocc in order to get coverage reports.

Most gocc test cases share the same structure, here is the common flow diagram:

```  
  (1)                                       (2)                                     
  (wait_profile_backend "xxx" &) --> wait ci-sync exist --> (goc profile -o filtered.cov)--
                                            |                                             |
                                            |                                             |
                                            | (4)                                         |
  --(gocc --debugcisyncfile ci-sync) --> finish, write ci-sync --> sleep 5; exit          |(5)
  |                                                                        |              |
  | (3)                                             (6)                    |              |
  |------------------------>(goc server &) --------------------------------|              |
                                  |                                                       |
                                  ---------------------------------------------------------
```

1. start the `wait_profile_backend` in the background.
2. `wait_profile_backend` will block until the file `ci-sync.bak` exists.
3. the covered `gocc` subcommand start and register to the `goc server`.
4. the covered `gocc` subcommand run its own logic until finish, as we add the `--debugcisyncfile ci-sync` flag, it will write a file called `ci-sync`, and wait 5 seconds.
5. `wait_profile_backend` continue to run, and try to get the profile from the `goc server`.
6. the `goc server` finally call the http API to get the `gocc` profile.
