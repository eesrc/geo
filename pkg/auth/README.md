# Auth

Auth is a simple library for adding known OAuth2 providers along with nonce, session creation and session handling.

## Authenticator

The authenticator is the parent of one or more providers which it shares its sessions with. It distributes http handlers for unauthorized and server errors to the provider, as well as handles the storage of the sessions. The store is an implementation of the `session-store` which by default supports sqlite3 and postgres.

## Provider

A provider is simply something that provides a way of authenticating a user and aggregating a `Profile` based on the available data for the provider. An `OAuth2Provider` has a corresponding Config object which needs to be filled by various optional fields depending on the provider. The provider also has the job of serving the auth endpoints, ie the OAuth2 callback endpoint. While fully configurable, the defaults allow for most users to just use the sane the providers with minimal configuration.

## Configuration

All config objects are primed with the [Params](<[http](https://github.com/exploratoryengineering/params)>) library, making it easy to configure your provider by just putting the config in your `Params`-init.
