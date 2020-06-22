# geo - a server for processing GPS tracking data

`geo` is a multipurpose tracking server with capabilities such as pub/sub, geofencing and storage of tracking data.

## Running geo

To run a local version of Geo you can run the `main.go` under `cmd/geo`, or build and use `geo` as a binary.

### Default configuration

Running `geo` without parameters it will run the geo server with sane defaults. It will populate an in-memory db with sqlite and expose the API on port `8877` along with the NATS streaming-server on `4222`.

### Running with flags

The binary `geo` also takes several flags which will configure the behaviour of the geo-server.

## Components

### Tria

Tria is the library which is responsible for handling GeoJSON data in which it triangulates and returns shapedata which in turn has several geo-functions for further use.

### Server

The server exposes a set of REST APIs for interacting and manipulating collections, trackers, shapecollections and subscriptions. It uses Tria to provide several functionalities such as tracker subscriptions and geofencing. The server exposes the API on port `8877`.

#### NATS

The server embeds a NATS streaming-server which is a eventbus available on port `4222`.

## Benchmarking

### Go benchmarks

Several performance critical files contain test benchmarks to ensure that the functions are performant enough.
