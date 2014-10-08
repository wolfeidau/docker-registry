# Docker Registry

This is a modified version of the golang implementation of the docker registry which was cloned from the docker contributions folder.

# Requirements

The service relies on a bunch of environment variables to operate.

```
    export REGISTRY_DATA=/data/docker       # where all the files are stored
    export REGISTRY_NAMESPACE=wolfeidau     # used in the docker URL similiar to your dockerhub user
    export REGISTRY_PASS="SETTHISNOW"       # global password used to log in to the registry
    export REGISTRY_SECRET="SETTHISNOW"     # secret for generating sessions
```    

# TODO

* Implement logins using something other than a single password
* Move to using JWT for sessions.
