# Media Service APIs
The Media service APIs handle all the media's domain operations.

They contain multiple layered APIs such as Public, Private and Admin API. 

It uses a gRPC and HTTP Sever to expose the APIs.

Alexandria is currently licensed under the MIT license.

## Endpoints
| Method     |     HTTP Mapping             |  HTTP Request body |  HTTP Response body |
|------------|:----------------------------:|:------------------:|:-------------------:|
| **List**   |  GET /collection-URL         |   N/A              |   Resource* list    |
| **Get**    |  GET /resource-URL           |   N/A              |   Resource*         |
| **Create** |  POST /collection-URL        |   Resource         |   Resource*         |
| **Update** |  PUT or PATCH /resource-URL  |   Resource         |   Resource*         |
| **Delete** |  DELETE /resource-URL        |   N/A              |   protobuf.empty    |

## Contribution
Alexandria is an open-source project, that means everyone’s help is appreciated.

If you'd like to contribute, please look at the [Go Contribution Guidelines](https://github.com/maestre3d/alexandria/tree/master/docs/GO_CONTRIBUTION.md).

[Click here](https://github.com/maestre3d/alexandria/tree/master/docs) if you're looking for our docs about engineering, Alexandria API, etc.

## Maintenance
- Main maintainer: [maestre3d](https://github.com/maestre3d)
