# Author Service API
The Author service API handles all the author's domain operations.

They contain multi-layered APIs such as Public, Private and Admin API. 

It uses a gRPC and HTTP Sever to expose the APIs.

Alexandria is currently licensed under the MIT license.

## Endpoints
| Method              |     HTTP Mapping                    |  HTTP Request body |  HTTP Response body    |
|---------------------|:-----------------------------------:|:------------------:|:----------------------:|
| **List**            |  GET /author                        |   N/A              |   Author* list         |
| **Create**          |  POST /author                       |   Author           |   Author*              |
| **Get**             |  GET /author/{author-id}            |   N/A              |   Author*              |
| **Update**          |  PUT or PATCH /author/{author-id}   |   Author           |   Author*              |
| **Delete**          |  DELETE /author/{author-id}         |   N/A              |   protobuf.empty/{}    |
| **Restore/Active**  |  PATCH /admin/author/{author-id}    |   N/A              |   protobuf.empty/{}    |
| **HardDelete**      |  DELETE /admin/author/{author-id}   |   N/A              |   protobuf.empty/{}    |

### Accepted Queries
The list method accepts multiple queries to make data fetching easier for everyone.

The following fields are accepted by our service.
- page_token = string
- page_size = int32 (min. 1, max. 100)
- query = string
- timestamp = boolean
- show_disabled = boolean
- owner_id = string (user from author's owners pool)
- ownership_type = string (public, private)

## Contribution
Alexandria is an open-source project, that means everyone’s help is appreciated.

If you'd like to contribute, please look at the [Go Contribution Guidelines](https://github.com/maestre3d/alexandria/tree/master/docs/GO_CONTRIBUTION.md).

[Click here](https://github.com/maestre3d/alexandria/tree/master/docs) if you're looking for our docs about engineering, Alexandria API, etc.

## Maintenance
- Main maintainer: [maestre3d](https://github.com/maestre3d)
