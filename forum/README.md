# Forum Microservice

## API First
This Module is using the API First approach to create a REST API. The API is defined in the `api` folder. The API is defined in the [OpenAPI 3.0](https://swagger.io/specification/) format. 
The API is in the `api/forum.yaml` file. The API is defined in the [OpenAPI 3.0](https://swagger.io/specification/) format.

To generate as much boilerplate code from the specification as possible, the tool [oapi-codegen](https://github.com/deepmap/oapi-codegen) is used.

If changes are made on the types in the specification, the types can be generated with the following command:
```shell
oapi-codegen -config ./api/types.cfg.yaml ./api/forum.yaml
```

If changes are made on the routes in the specification, the routes and server-code can be generated with the following command:
```shell
oapi-codegen -config ./api/server.cfg.yaml ./api/forum.yaml
```

## Structure
To be described

## Build
```shell
go build
```

## Start
```shell
./forum
```