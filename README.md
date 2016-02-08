# Duplex Hub

Duplex Hub allows Duplex services to be securely published on the web. It currently only works for Duplex services using WebSocket as the transport and JSON as the codec. Run on Heroku, then connect with WebSocket passing auth tokens. Hand that socket connection to your local Duplex RPC. Now anybody with proper credentials can connect via WebSocket to the gateway and interact with your private Duplex services.

TODO: improve this description

## Running the Hub

Deploy on Heroku...

## Publishing Duplex Services

First connect to the hub over HTTPS. Use a path that you'd like to use as a public endpoint. You need to authenticate by passing a `token` secret as a query parameter that the gateway was configured with. Also pass a `secret` query parameter that will be used to authenticate clients connecting to your endpoint.

Now upgrade to WebSocket. Over WebSocket, perform the Duplex handshake. This connection is now like any other Duplex connection and can send and receive requests and replies. Clients connecting to the hub will have their own connection, but their messages will be multiplexed over this single connection.

## Using Duplex Services

Given a known endpoint that services are exposed on, you can connect to that endpoint like a regular Duplex peer over WebSocket transport. You just have to connect with HTTPS passing a `secret` query parameter. That's it!

## Using Services via HTTP (future)

You can also perform HTTP POST requests against subpaths of the endpoint. The subpath will be used as the method to make a request against. Your body will be used as the request payload. The response will be the reply payload as JSON. These requests also require the `secret` query parameter.

## TODO

 * run on heroku
 * more tests
   * can't upstream without token+secret
   * can't client without secret
   * multiple endpoints
   * ERRORS client doesn't exist any more, etc
