Building and Running Docker Images
----------------------------------

## Building images

To easily build any of the Docker files in this directory, we can use the `build.sh` script.

Usage:
```
Usage: build.sh [OPTIONS]

OPTIONS:
  -h                                          Print the help output
  -build [all, willow, limiter, locker]       Build a specific docker image. If this flag is not provided the default 'all' will be used
```

Or if you prrefer, everything can be built with `docker compose`:
```
docker compose build [OPTIONAL NAME OF SERVICE]
```

## Running images

* Limiter in this example is exposed over http on port 8082
* Locker in this example is exposed over http on port 8083

All services have a `http://127.0.0.1:[port]/docs` api that can be viewed for details service instructions

To run the the `Limiter` and `Locker` services through `docker compose`:
```
docker compose up
``` 

To run them directly without `docker compose` requires a bit more setup as the `Limiter` service 
relies on `Locker`. So we need to create a custom network for the 2 services to run in.
```
# setup the neetwork
docker network create willow

# run locker
docker run -p 8083:8083 --network willow --name locker locker:local-latest

# run limiter
docker run -p 8082:8082 --network willow -name limiter limiter:local-latest /bin/bash -c "limiter -limiter-insecure-http -log-level=debug -limiter-locker-url http://locker:8083"
```

**NOTE**
At the moment, `Willow` is not up and running as I have not made it run through HTTP, it currently only runs with HTTPS
and thats a bit annoying as a 1st pass to try out. So for now, there is no willow service in the compose file. Also there
is a bunch of development work going on to make `Willow` use the `Limiter` service which will be the driving factors behind
this feature.