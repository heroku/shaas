# shaas
Shell as a Service

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/heroku/shaas)

## Overview
API to inspect and execute scripts in the server's environment.

**This is obviously a *really bad idea* on a server that you care about, but this is a convenience for testing purposes only. This offers no protection whatsoever for the server. This makes the server's entire file system accessible to clients. Please use with great caution.**

## Usage

Summary of endpoint behavior for all path, method, and protocol combinations:

|           |                 POST                  |         GET         |                      WebSocket                      |
|-----------|---------------------------------------|---------------------|-----------------------------------------------------|
| File      | runs path in context of its directory | downloads path      | interactively runs path in context of its directory |
| Directory | runs body in context of path          | lists files in path | runs interactive shell in context of path           |

### Executing Commands

To execute a command in the context of a given directory on the server, simply `POST` the command with the directory as the URL path. The command runs with CGI environment variables. For example, running `pwd` in the directory `/app/views` returns the path in the response:

```
$ curl http://shaas.example.com/app/views -i -X POST -d 'pwd'
HTTP/1.1 200 OK
Date: Tue, 21 Apr 2015 17:22:07 GMT
Content-Type: text/plain; charset=utf-8
Transfer-Encoding: chunked

/app/views
```

This is the most versatile endpoint. The functionality of all the other endpoints could be achieved with a `POST` to a directory path, but are offered as a convenience.

### Executing Scripts

To execute a script on the server, simply `POST` the script path as the URL path. The script runs with CGI environment variables. For example, to run an executable script at `/app/bin/migrate`:

```
$ curl http://shaas.example.com/app/bin/migrate -i -X POST -d 'input to script'
HTTP/1.1 200 OK
Date: Tue, 21 Apr 2015 17:22:07 GMT
Content-Type: text/plain; charset=utf-8
Transfer-Encoding: chunked

migration complete
```

### Interactive Sessions

By accessing the endpoints above via WebSockets, the commands are run interactively. If the path is a directory, an interactive `bash` session is started in that directory. If the path is a script, it is run in an interactive session. For example, using the [wssh](https://github.com/progrium/wssh) client:

```
$ wssh ws://shaas.example.com/app
/app $ echo 'hello'
echo 'hello'
hello
```

### Listing a Directory

Directories are listed in JSON format for easy parsing:

```
$ curl http://shaas.example.com/app -i -X GET
HTTP/1.1 200 OK
Content-Type: application/json
Date: Tue, 21 Apr 2015 17:26:53 GMT
Content-Length: 1020

{
  "views": {
    "size": 11,
    "type": "d",
    "permission": 493,
    "updated_at": "2015-04-20T21:38:49-07:00"
  },
  "README.md": {
    "size": 1924,
    "type": "-",
    "permission": 420,
    "updated_at": "2015-04-21T10:27:37-07:00"
  }
}
```

If viewing the directory in a browser (or any client with a `html` in the `Accept` header), the listing will be returned in HTML:

```
$ curl http://shaas.example.com/app -i -X GET -H 'Accept: text/html'
HTTP/1.1 200 OK
Content-Type: text/html
Date: Tue, 21 Apr 2015 17:46:58 GMT
Content-Length: 185

<ul>
    <li><a href='views'>/views</a></li>
    <li><a href='README.md'>README.md</a></li>
</ul>
}
```

To list a directory in plain text, use POST with the `ls` command and options of your choice:

```
$ curl http://shaas.example.com/app -i -X POST -d 'ls -lA'
HTTP/1.1 200 OK
Date: Tue, 21 Apr 2015 17:35:43 GMT
Content-Type: text/plain; charset=utf-8
Transfer-Encoding: chunked

total 64
drwxr-xr-x  12 user  454177323   408 Apr 21 10:35 views
-rw-r--r--   1 user  454177323  2268 Apr 21 10:35 README.md
```

### Downloading a File

Files are returned in their native format:

```
$ curl http://shaas.example.com/app/images/logo.jpeg -i -X GET
HTTP/1.1 200 OK
Date: Tue, 21 Apr 2015 17:31:45 GMT
Content-Type: image/jpeg

<BINARY DATA>
```

## Deployment

    $ heroku create --buildpack https://github.com/heroku/heroku-buildpack-go.git
    $ git push heroku master
    
... or just:

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/heroku/shaas)
