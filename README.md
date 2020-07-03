# goc
[![Go Report Card](https://goreportcard.com/badge/github.com/qiniu/goc)](https://goreportcard.com/report/github.com/qiniu/goc)
![](https://github.com/qiniu/goc/workflows/ut-check/badge.svg)
![](https://github.com/qiniu/goc/workflows/style-check/badge.svg)
![](https://github.com/qiniu/goc/workflows/e2e%20test/badge.svg)
[![codecov](https://codecov.io/gh/qiniu/goc/branch/master/graph/badge.svg)](https://codecov.io/gh/qiniu/goc)
[![GoDoc](https://godoc.org/github.com/qiniu/goc?status.svg)](https://godoc.org/github.com/qiniu/goc)

goc is a comprehensive coverage testing system for The Go Programming Language, especially for some complex scenarios，like system testing code coverage collection and
accurate testing.

Enjoy, Have Fun!
![Demo](docs/images/intro.gif)

## Installation
To install goc tool, you need to install Go first (**version 1.11+ is required**), then:

```go get -u github.com/qiniu/goc```

## Examples
You can use goc tool in many scenarios.

### Code Coverage Collection for your Golang System Tests
Goc can collect code coverages at runtime for your long-run golang applications. To do that, normally just need three steps:

1. use `goc server` to start a service registry center:
    ```
    ➜  simple-go-server git:(master) ✗ goc server 
    ```
2. use `goc build` to build the target service, and run the generated binary. Here let's take the [simple-go-server](https://github.com/CarlJi/simple-go-server) project as example:
    ```
    ➜  simple-go-server git:(master) ✗ goc build .
    ... // omit logs
    ➜  simple-go-server git:(master) ✗ ./simple-go-server  
    ```
3. use `goc profile` to get the code coverage profile of the started simple server above:
    ```
    ➜  simple-go-server git:(master) ✗ goc profile
    mode: atomic
    enricofoltran/simple-go-server/main.go:30.13,48.33 13 1
    enricofoltran/simple-go-server/main.go:48.33,50.3 1 0
    enricofoltran/simple-go-server/main.go:52.2,65.12 5 1
    enricofoltran/simple-go-server/main.go:65.12,74.46 7 1
    enricofoltran/simple-go-server/main.go:74.46,76.4 1 0
    ...   
    ```
### Code Coverage Collection when your golang application run in Docker
Goc can collect code coverages at runtime for your golang applications run in Docker. To do that, normally just need three steps:
1. use `goc server` to start a service registry center in host machine:
    ```
    ➜  root ✗ goc server --port :8110 
    ```
2. use `goc build` to build the target service with host machine address and fixed agent port:
    ```
    ➜  simple-go-server git:(master) ✗ goc build --center=http://${host-machine-ip}:8110 --agentport=:8111
    ... // omit logs
    ```   
3. use generated simple binary build [Docker images](https://hub.docker.com/repository/docker/memoryliu/goc-simple) , then use `docker run` start container with fixed agent port:
    ```
    // pull simple image
    ➜  root ✗ docker pull memoryliu/goc-simple:v1
   // docker run with port mapping
    ➜  root ✗ docker run -it -p 8111:8111 memoryliu/goc-simple:v1
   // get container id
    ➜  root ✗ docker ps -a | grep goc
    4d3b5698a27a        memoryliu/goc-simple:v1   "/simple-go-server"    0.0.0.0:8111->8111/tcp 
    ```

4. use `docker exec` execute goc and output code coverages:
    ```
    ➜  root ✗ docker exec 4d3b5698a27a /goc profile --center=http://${host-machine-ip}:8110
    mode: count
    enricofoltran/simple-go-server/main.go:30.13,48.33 13 3
    enricofoltran/simple-go-server/main.go:48.33,50.3 1 0
    enricofoltran/simple-go-server/main.go:52.2,65.12 5 3
    enricofoltran/simple-go-server/main.go:65.12,74.46 7 3
    enricofoltran/simple-go-server/main.go:74.46,76.4 1 0
    enricofoltran/simple-go-server/main.go:77.3,77.14 1 0
    enricofoltran/simple-go-server/main.go:80.2,82.79 3 3
    ```


## RoadMap
- [x] Support code coverage collection for system testing.
- [x] Support code coverage counters clear for the services under test at runtime.
- [ ] Support develop mode towards accurate testing.
- [ ] Support code coverage diff based on Pull Request.
- [ ] Optimize the performance costed by code coverage counters.

## Contributing
We welcome all kinds of contribution, including bug reports, feature requests, documentation improvements, UI refinements, etc.

Thanks to all [contributors](https://github.com/qiniu/goc/graphs/contributors)!!

## License
Goc is released under the Apache 2.0 license. See [LICENSE.txt](https://github.com/qiniu/goc/blob/master/LICENSE)