# Simplex Gateway

Simplex Gateway allows private Simplex services to be securely exposed on the web. It uses websocket as the transport and currently assumes JSON as the codec. Run on Heroku, then connect with websocket after authenticating with HTTP. Hand that socket connection to your local Simplex RPC. Now anybody with proper credentials can connect via websocket to the gateway and interact with your private Simplex services.

TODO: improve this description

## Running the Gateway

Deploy on Heroku...

## Exposing Simplex Services

First connect to the gateway over HTTPS. Use a path that you'd like to use as a public endpoint. You need to authenticate by passing a `token` secret as a query parameter that the gateway was configured with. Also pass a `secret` query parameter that will be used to authenticate clients connecting to your endpoint.

Now upgrade to Websocket. Over Websocket, perform the Simplex handshake. This connection is now like any other Simplex connection and can send and receive requests and replies. Clients connecting to the gateway will have their own connection, but their messages will be multiplexed over this single connection.

## Using Simplex Services

Given a known endpoint that services are exposed on, you can connect to that endpoint like a regular Simplex peer over Websocket transport. You just have to connect with HTTPS passing a `secret` query parameter. That's it!

## Using Services via HTTP (future)

You can also perform HTTP POST requests against subpaths of the endpoint. The subpath will be used as the method to make a request against. Your body will be used as the request payload. The response will be the reply payload as JSON. These requests also require the `secret` query parameter.

## TODO

 * run on heroku
 * more tests
   * can't upstream without token+secret
   * can't client without secret
   * multiple endpoints
   * ERRORS client doesn't exist any more, etc
