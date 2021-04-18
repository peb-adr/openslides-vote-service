# OpenSlides Vote Service

The Vote Service is part of the OpenSlides environments. It handles the votes
for an electonic poll.


## Install and Start

TODO


## Example Request with CURL

### Start a Poll

To start a poll a POST request has to be send to the create-url. The body has to
be a valid poll-config.

To send the same request twice is ok. But to send a different poll-config is an
error.

```
curl localhost:9013/internal/vote/create?pid=1 -d '{"content_object_id":"motion/2", "backend":"fast"}'
```


### Send a Vote

A vote-request is a post request with the ballot as body. Only logged in users
can vote. The body has to be valid json. For example for the value 'Y' you have
to send `"Y"`.

This handler is not idempotent. If the same user sends the same data twice, it
is an error.

```
curl localhost:9013/system/vote?pid=1 -d '"Y"'
```


### Stop the Poll

With the stop request a poll is stopped and the vote values are returned. The
stop request is a POST request without a body.

A stop request can be send many times and will return the same data again.

```
curl localhost:9013/internal/vote/stop?pid=1 -d ''
```


### Clear the poll

After a vote was stopped and the data is successfully stored in the datastore, a
clear request should be used to remove the data from the vote service. This is
especially important on fast votes to remove the mapping between the user id and
the vote. The clear requet is idempotent.

```
curl localhost:9013/internal/vote/clear?pid=1 -d ''
```


## Configuration

### Environment variables

The Service uses the following environment variables:

* `VOTE_HOST`: The device where the service starts. The default is am empty
  string which starts the service on any device.
* `VOTE_PORT`: The port the vote service listens on. The default is `9012`. 
* `DATASTORE_READER_HOST`: Host of the datastore reader. The default is
  `localhost`.
* `DATASTORE_READER_PORT`: Port of the datastore reader. The default is `9010`.
* `DATASTORE_READER_PROTOCOL`: Protocol of the datastore reader. The default is
  `http`.
* `MESSAGING`: Sets the type of messaging service. `fake`(default) or
  `redis`.
* `MESSAGE_BUS_HOST`: Host of the redis server. The default is `localhost`.
* `MESSAGE_BUS_PORT`: Port of the redis server. The default is `6379`.
* `REDIS_TEST_CONN`: Test the redis connection on startup. Disable on the cloud
  if redis needs more time to start then this service. The default is `true`.
* `VOTE_REDIS_HOST`: Host of the redis used for the fast backend and the vote
  config. Default is `localhost'.
* `VOTE_REDIS_PORT`: Port of the redis host. Default is `6379`.
* `AUTH`: Sets the type of the auth service. `fake` (default) or `ticket`.
* `AUTH_HOST`: Host of the auth service. The default is `localhost`.
* `AUTH_PORT`: Port of the auth service. The default is `9004`.
* `AUTH_PROTOCOL`: Protocol of the auth servicer. The default is `http`.
* `OPENSLIDES_DEVELOPMENT`: If set, the service starts, even when secrets (see
  below) are not given. The default is `false`.


### Secrets

Secrets are filenames in `/run/secrets/`. The service only starts if it can find
each secret file and read its content. The default values are only used, if the
environment variable `OPENSLIDES_DEVELOPMENT` is set.

* `auth_token_key`: Key to sign the JWT auth tocken. Default `auth-dev-key`.
* `auth_cookie_key`: Key to sign the JWT auth cookie. Default `auth-dev-key`.
