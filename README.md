# shaas
Shell as a Service

## Overview
API to inspect and execute scripts in a server's environment via HTTP and WebSockets.

**This is obviously a *really bad idea* on a server that you care about, but this is a convenience for testing purposes only. This offers no protection whatsoever for the server. This makes the server's entire file system accessible to clients. Please use with great caution.**

## Running

Because this application gives clients full access to the server, it is highly recommended to run it inside of some kind of containerized environment, such as [Heroku](http://www.heroku.com) or [Docker](https://www.docker.com/). Even in a containerized environment, you may wish to set a username and password, for use via HTTP basic authentication, by setting `BASIC_AUTH=user:password` in the environment before starting. To only allow `GET` requests and disallow websockets, set `READ_ONLY` in the environment.

### Heroku

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/heroku/shaas)

### Docker

Running with [Docker Compose](https://docs.docker.com/compose):

    $ docker-compose up -d
    $ curl http://localhost:5000/

## Usage

Summary of endpoint behavior for all path, method, and protocol combinations:

|           |                 POST                  |         GET         |      PUT/APPEND       |                      WebSocket                      |
|-----------|---------------------------------------|---------------------|-----------------------|-----------------------------------------------------|
| File      | runs path in context of its directory | downloads path      | uploads body to path  | interactively runs path in context of its directory |
| Directory | runs body in context of path          | lists files in path | n/a                   | runs interactive shell in context of path           |

### Executing Commands

To execute a command in the context of a given directory on the server, simply `POST` the command with the directory as the URL path. For example, running `pwd` in the directory `/usr/bin` returns the path in the response:

    $ curl http://shaas.example.com/usr/bin -i -X POST -d 'pwd'
    HTTP/1.1 200 OK
    Date: Tue, 21 Apr 2015 17:22:07 GMT
    Content-Type: text/plain; charset=utf-8
    Transfer-Encoding: chunked

    /usr/bin

This is the most versatile endpoint. The functionality of all the other endpoints could be achieved with a `POST` to a directory path, but are offered as a convenience.

### Executing Scripts

To execute a script on the server, simply `POST` the script path as the URL path and any input to the script in the body. For example, to find the factors of the number 24:

    $ curl http://shaas.example.com/usr/bin/factor -i -X POST -d '24'
    HTTP/1.1 200 OK
    Server: Cowboy
    Connection: keep-alive
    Date: Fri, 15 May 2015 16:40:08 GMT
    Content-Type: text/plain; charset=utf-8
    Transfer-Encoding: chunked
    Via: 1.1 vegur
    
    24: 2 2 2 3
    
Because `/usr/bin` is on the `PATH`, this could also be run with just the command in the body:

    $ curl http://shaas.example.com/ -i -X POST -d 'factor 24'
    HTTP/1.1 200 OK
    Server: Cowboy
    Connection: keep-alive
    Date: Fri, 15 May 2015 16:45:43 GMT
    Content-Type: text/plain; charset=utf-8
    Transfer-Encoding: chunked
    Via: 1.1 vegur
    
    24: 2 2 2 3

### CGI Environment Variables

All commands and scripts are automatically run with [CGI](http://en.wikipedia.org/wiki/Common_Gateway_Interface) environment variables for access to HTTP headers, query parameters, and other metadata:
    
    $ curl http://shaas.example.com/ -X POST -d 'env | sort'
    CONTENT_LENGTH=10
    CONTENT_TYPE=application/x-www-form-urlencoded
    GATEWAY_INTERFACE=CGI/1.1
    HTTP_ACCEPT=*/*
    HTTP_CONNECTION=close
    HTTP_CONNECT_TIME=5
    HTTP_CONTENT_LENGTH=10
    HTTP_CONTENT_TYPE=application/x-www-form-urlencoded
    HTTP_HOST=shaas.example.com
    HTTP_TOTAL_ROUTE_TIME=0
    HTTP_USER_AGENT=curl/7.37.1
    HTTP_VIA=1.1 vegur
    HTTP_X_FORWARDED_FOR=73.170.209.186
    HTTP_X_FORWARDED_PORT=80
    HTTP_X_FORWARDED_PROTO=http
    HTTP_X_REQUEST_ID=9003884b-310a-4095-8ff1-0894494aff75
    HTTP_X_REQUEST_START=1429846020992
    PATH_INFO=/
    PWD=/
    QUERY_STRING=
    REMOTE_ADDR=10.216.205.205:30916
    REMOTE_HOST=10.216.205.205:30916
    REQUEST_METHOD=POST
    REQUEST_URI=/
    SCRIPT_FILENAME=/
    SCRIPT_NAME=/
    SERVER_NAME=shaas.example.com
    SERVER_PORT=23389
    SERVER_PROTOCOL=HTTP/1.1
    SERVER_SOFTWARE=go
    SHLVL=1
    _=/usr/bin/env

### Interactive Sessions

By accessing the endpoints above via WebSockets, the commands are run interactively. If the path is a directory, an interactive `bash` session is started in that directory. If the path is a script, it is run in an interactive session. For example, using the [wssh](https://github.com/progrium/wssh) client:

    $ wssh ws://shaas.example.com/
    / $ echo 'hello'
    echo 'hello'
    hello

### Listing a Directory

Directories are listed in JSON format for easy parsing:


    $ curl http://shaas.example.com/usr -i -X GET
    HTTP/1.1 200 OK
    Server: Cowboy
    Connection: keep-alive
    Content-Type: application/json
    Date: Fri, 15 May 2015 16:52:29 GMT
    Content-Length: 996
    Via: 1.1 vegur
    
    {
      "bin": {
        "size": 36864,
        "type": "d",
        "permission": 493,
        "updated_at": "2015-03-20T09:28:58.547556085Z"
      },
      "games": {
        "size": 4096,
        "type": "d",
        "permission": 493,
        "updated_at": "2014-04-10T22:12:14Z"
      }
    }

If viewing the directory in a browser (or any client with a `html` in the `Accept` header), the listing will be returned in HTML:

    $ curl http://shaas.example.com/usr -i -X GET -H 'Accept: text/html'
    HTTP/1.1 200 OK
    Content-Type: text/html
    Date: Tue, 21 Apr 2015 17:46:58 GMT
    Content-Length: 185

    <ul>
        <li><a href='bin'>/bin</a></li>
        <li><a href='games'>/games</a></li>
    </ul>

To list a directory in plain text, use POST with the `ls` command and options of your choice:

    $ curl http://shaas.example.com/usr -i -X POST -d 'ls -lA'
      HTTP/1.1 200 OK
      Server: Cowboy
      Connection: keep-alive
      Date: Fri, 15 May 2015 16:54:28 GMT
      Content-Type: text/plain; charset=utf-8
      Transfer-Encoding: chunked
      Via: 1.1 vegur
      
      total 72
      drwxr-xr-x   2 root root 36864 Mar 20 09:28 bin
      drwxr-xr-x   2 root root  4096 Apr 10  2014 games

### Downloading a File

Files are returned in their native format:

    $ curl http://shaas.example.com/var/logs/server.log -i -X GET
    HTTP/1.1 200 OK
    Date: Tue, 21 Apr 2015 17:31:45 GMT
    Content-Type: plain/text

    ...

### Uploading a File

`PUT` creates or replaces a file with the request body:

    $ curl http://shaas.example.com/var/logs/server.log -i -X PUT --data-binary 'hello 1'
    HTTP/1.1 200 OK
    Date: Tue, 28 Mar 2017 09:13:05 GMT
    Content-Length: 0
    Content-Type: text/plain; charset=utf-8

    $ curl http://shaas.example.com/var/logs/server.log -i -X PUT --data-binary 'hello 2'
    HTTP/1.1 200 OK
    Date: Tue, 28 Mar 2017 09:13:05 GMT
    Content-Length: 0
    Content-Type: text/plain; charset=utf-8

    $ curl localhost:5000/var/logs/server.log -i
    HTTP/1.1 200 OK
    Content-Length: 7
    Date: Tue, 28 Mar 2017 09:13:38 GMT
    Content-Type: text/plain; charset=utf-8

    hello 2

`APPEND` creates or appends a file with the request body:

    $ curl http://shaas.example.com/var/logs/server.log -i -X APPEND --data-binary 'hello 1'
    HTTP/1.1 200 OK
    Date: Tue, 28 Mar 2017 09:13:05 GMT
    Content-Length: 0
    Content-Type: text/plain; charset=utf-8

    $ curl http://shaas.example.com/var/logs/server.log -i -X APPEND --data-binary 'hello 2'
    HTTP/1.1 200 OK
    Date: Tue, 28 Mar 2017 09:13:05 GMT
    Content-Length: 0
    Content-Type: text/plain; charset=utf-8

    $ curl localhost:5000/var/logs/server.log -i
    HTTP/1.1 200 OK
    Content-Length: 14
    Date: Tue, 28 Mar 2017 11:56:43 GMT
    Content-Type: text/plain; charset=utf-8

    hello-1hello-2

## Overriding HTTP Methods

Because not all clients support all HTTP methods, particularly `PUT` and the custom `APPEND` method, the method can alternatively be overridden with the `_method` query parameter alongd with the `POST` method. For example, the following are equivalent:

    $ curl http://shaas.example.com/var/logs/server.log -i -X APPEND --data-binary 'hello 1'

    $ curl http://shaas.example.com/var/logs/server.log?_method=APPEND -i -X POST --data-binary 'hello 1'

## Testing

Due to the nature of this application and the access it has to the host machine, testing is done functionality within a [Docker](https://www.docker.com) container. To run tests, be sure Docker is running, [Docker Compose](https://docs.docker.com/compose) is installed, and run:

    $ go test -v ./... -ftest
