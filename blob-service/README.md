# Blob Service API
The Blob service API handles all the author's domain operations.

It contains multi-layered APIs such as Public, Private and Admin API. 

It uses gRPC and HTTP communication protocols to expose its APIs.

Alexandria is currently licensed under the MIT license.

## Endpoints

| Method              |     HTTP Mapping                          |  HTTP Request body  |  HTTP Response body        |
|---------------------|:-----------------------------------------:|:-------------------:|:--------------------------:|
| **Get**             |  GET /blob/author/{author-id}             |   N/A               |   Blob*                    |
| **Store**           |  POST /private/blob/author/{author-id}    |   File              |   Blob*                    |
| **Delete**          |  DELETE /private/blob/author/{author-id}  |   N/A               |   protobuf.empty/{}        |
| **Get**             |  GET /blob/media/{media-id}               |   N/A               |   Blob*                    |
| **Store**           |  POST /private/blob/media/{media-id}      |   File              |   Blob*                    |
| **Delete**          |  DELETE /private/blob/media/{media-id}    |   N/A               |   protobuf.empty/{}        |
| **Get**             |  GET /blob/user/{user-id}                 |   N/A               |   Blob*                    |
| **Store**           |  POST /private/blob/user/{user-id}        |   File              |   Blob*                    |
| **Delete**          |  DELETE /private/blob/user/{user-id}      |   N/A               |   protobuf.empty/{}        |

### Accepted Queries
The list method accepts multiple queries to make data fetching easier for everyone.

The following fields are valid for our service.
- page_token = string
- page_size = int32 (min. 1, max. 100)
- query = string
- filter_by = string (id, timestamp or popularity by default)
- sort = string (asc or desc)
- show_disabled = boolean

## Contribution
Alexandria is an open-source project, that means everyone’s help is appreciated.

If you'd like to contribute, please look at the [Go Contribution Guidelines](https://github.com/maestre3d/alexandria/tree/master/docs/GO_CONTRIBUTION.md).

[Click here](https://github.com/maestre3d/alexandria/tree/master/docs) if you're looking for our docs about engineering, Alexandria API, etc.

## Maintenance
- Main maintainer: [maestre3d](https://github.com/maestre3d)
