# Identity Service API
The Identity service API handles all the user's domain operations.

It contains multi-layered APIs such as Public, Private and Admin API. 

It uses gRPC and HTTP communication protocols to expose its APIs.

The Identity API v1 is using Amazon Web Service's Cognito as user pool and identity federation manager.

Alexandria is currently licensed under the MIT license.

## Endpoints
| Method              |     HTTP Mapping                        |  HTTP Request body  |  HTTP Response body |
|---------------------|:---------------------------------------:|:-------------------:|:-------------------:|
| **List**            |  GET /admin/user                        |   N/A               |   User* list        |
| **Get**             |  GET /user/{author-id}                  |   N/A               |   User*             |
| **Update**          |  PUT or PATCH /private/user/{user-id}   |   User              |   User*             |
| **Delete**          |  DELETE /private/user/{user-id}         |   N/A               |   protobuf.empty/{} |
| **Restore/Active**  |  PATCH /private/userr/{user-id}         |   N/A               |   protobuf.empty/{} |
| **HardDelete**      |  DELETE /admin/user/{user-id}           |   N/A               |   protobuf.empty/{} |

### Accepted Queries
The list method accepts multiple queries to make data fetching easier for everyone.

The following fields are accepted by our service.
- page_token = string
- page_size = int32 (min. 1, max. 100)
- query = string
- filter_by = string (id or timestamp)
- sort = string (asc or desc)
- show_disabled = boolean


## Contribution
Alexandria is an open-source project, that means everyone’s help is appreciated.

If you'd like to contribute, please look at the [Go Contribution Guidelines](https://github.com/maestre3d/alexandria/tree/master/docs/GO_CONTRIBUTION.md).

[Click here](https://github.com/maestre3d/alexandria/tree/master/docs) if you're looking for our docs about engineering, Alexandria API, etc.

## Maintenance
- Main maintainer: [maestre3d](https://github.com/maestre3d)
