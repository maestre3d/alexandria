# Media Service API
The Media service API handles all the media's domain operations.

It contains multi-layered APIs such as Public, Private and Admin API. 

It uses gRPC and HTTP communication protocols to expose it's APIs.

Alexandria is currently licensed under the MIT license.

## Endpoints
| Method     |     HTTP Mapping                  |  HTTP Request body |  HTTP Response body   |
|------------|:---------------------------------:|:------------------:|:---------------------:|
| **List**   |  GET /media                       |   N/A              |   Media* list         |
| **Create** |  POST /media                      |   Media            |   Media*              |
| **Get**    |  GET /media/{media-id}            |   N/A              |   Media*              |
| **Update** |  PUT or PATCH /media/{media-id}   |   Media            |   Media*              |
| **Delete** |  DELETE /media/{media-id}         |   N/A              |   protobuf.empty/{}   |
| **Restore/Active**  |  PATCH /admin/media/{media-id}    |   N/A              |   protobuf.empty/{}    |
| **HardDelete**      |  DELETE /admin/media/{media-id}   |   N/A              |   protobuf.empty/{}    |

### Accepted Queries
The list method accepts multiple queries to make data fetching easier for everyone.

The following fields are accepted by our service.
- page_token = string
- page_size = int32 (min. 1, max. 100)
- search_query = string
- timestamp = boolean
- media_type = string
- author = string
- user = string

## Contribution
Alexandria is an open-source project, that means everyone’s help is appreciated.

If you'd like to contribute, please look at the [Go Contribution Guidelines](https://github.com/maestre3d/alexandria/tree/master/docs/GO_CONTRIBUTION.md).

[Click here](https://github.com/maestre3d/alexandria/tree/master/docs) if you're looking for our docs about engineering, Alexandria API, etc.

## Maintenance
- Main maintainer: [maestre3d](https://github.com/maestre3d)
