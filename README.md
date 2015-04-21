# shaas
Shell as a Service

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/heroku/shaas)

## Overview
REST API to shell out to server's environment. This is obviously a *really bad idea* on a server that you care about, but this is a convenience for testing servers that can only be accessed via HTTP.

## Usage

## Testing

    $ go test

## Deployment

    $ heroku create --buildpack https://github.com/heroku/heroku-buildpack-go.git
    $ git push heroku master
    
... or just:

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/heroku/shaas)
