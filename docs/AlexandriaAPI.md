# Alexandria API Design
To build a simple and comprehensible API, Alexandria crafted its API with help of Google’s Cloud API Design Guide.

In the following section, we define our API’s key concepts.

| **Type**                |  **Name**                            |
|-------------------------|:------------------------------------:|
| **Product**             |  **Alexandria** API                  |
| **Service**             |  alexandria.maestreapis.com          |
| **Package**             |  maestre.alexandria.v1               |
| **Interface**           |  maestre.alexandria.v1.ServiceName   |
| **Source Directory**    |  //maestre/alexandria/v1             |
| **API**                 |  alexandria                          |


## Resources (Endpoints)
_“In resource-oriented APIs, resources are named entities, and resource names are their identifiers. Each resource must have its own unique resource name. A collection is a special kind of resource that contains a list of sub-resources of identical type. The resource ID for a collection is called collection ID.” (Google, 2020)_

In the following list, we define our API’s resources and collection IDs.

### Identity API
| Method     |     HTTP Mapping                |  HTTP Request body |  HTTP Response body    |
|------------|:-------------------------------:|:------------------:|:----------------------:|
| **List**   |  GET /user                      |   N/A              |   User* list           |
| **Create** |  POST /user                     |   User             |   User*                |
| **Get**    |  GET /user/{user-id/username}   |   N/A              |   User*                |
| **Update** |  PUT or PATCH /user/{user-id}   |   User             |   User*                |
| **Delete** |  DELETE /user/{user-id}         |   N/A              |   protobuf.empty/{}    |

| Method      |     HTTP Mapping                |  HTTP Request body |  HTTP Response body    |
|-------------|:-------------------------------:|:------------------:|:----------------------:|
| **Auth**    |  POST /auth                     |   Credentials      |   JWTToken*            |
| **Refresh** |  GET /auth/refresh              |   Credentials      |   JWTToken*            |

_*Refresh uses Cookies, Set-Cookie:refresh_token_

_For more information, please [click here](https://hasura.io/blog/best-practices-of-using-jwt-with-graphql/)._

### Author API
| Method     |     HTTP Mapping                    |  HTTP Request body |  HTTP Response body    |
|------------|:-----------------------------------:|:------------------:|:----------------------:|
| **List**   |  GET /author                        |   N/A              |   Author* list         |
| **Create** |  POST /author                       |   Author           |   Author*              |
| **Get**    |  GET /author/{author-id}            |   N/A              |   Author*              |
| **Update** |  PUT or PATCH /author/{author-id}   |   Author           |   Author*              |
| **Delete** |  DELETE /author/{author-id}         |   N/A              |   protobuf.empty/{}    |


### Media API
| Method     |     HTTP Mapping                  |  HTTP Request body |  HTTP Response body   |
|------------|:---------------------------------:|:------------------:|:---------------------:|
| **List**   |  GET /media                       |   N/A              |   Media* list         |
| **Create** |  POST /media                      |   Media            |   Media*              |
| **Get**    |  GET /media/{media-id}            |   N/A              |   Media*              |
| **Update** |  PUT or PATCH /media/{media-id}   |   Media            |   Media*              |
| **Delete** |  DELETE /media/{media-id}         |   N/A              |   protobuf.empty/{}   |


### Category API
| Method     |     HTTP Mapping                                     |  HTTP Request body    |  HTTP Response body       |
|------------|:----------------------------------------------------:|:---------------------:|:-------------------------:|
| **List**   |  GET /category                                       |   N/A                 |   Category* list          |
| **Create** |  POST /category                                      |   Category            |   Category*               |
| **Get**    |  GET /category/{category-id}                         |   N/A                 |   Category*               |
| **Update** |  PUT or PATCH /category/{category-id}                |   Category            |   Category*               |
| **Delete** |  DELETE /category/{category-id}                      |   N/A                 |   protobuf.empty/{}       |

| Method     |     HTTP Mapping                                     |  HTTP Request body    |  HTTP Response body       |
|------------|:----------------------------------------------------:|:---------------------:|:-------------------------:|
| **Create** |  POST /category/{category-id}/author                 |   CategoryAuthor      |   CategoryAuthor*         |
| **List**   |  GET /category/{category-id}/author/{author-id}      |   N/A                 |   CategoryAuthor* list    |
| **Delete** |  DELETE /category/{category-id}/author/{author-id}   |   N/A                 |   protobuf.empty/{}       |

| Method     |     HTTP Mapping                                     |  HTTP Request body    |  HTTP Response body       |
|------------|:----------------------------------------------------:|:---------------------:|:-------------------------:|
| **Create** |  POST /category/{category-id}/media                  |   CategoryMedia       |   CategoryMedia*          |
| **List**   |  GET /category/{category-id}/media/{media-id}        |   N/A                 |   CategoryMedia* list     |
| **Delete** |  DELETE /category/{category-id}/media/{media-id}     |   N/A                 |   protobuf.empty/{}       |


### Blob API
| Method     |     HTTP Mapping                             |  HTTP Request body    |  HTTP Response body       |
|------------|:--------------------------------------------:|:---------------------:|:-------------------------:|
| **Create** |  POST /blob/user                             |   Multipart           |   UserImgURL*             |
| **Get**    |  GET /blob/user/{user-id/username}           |   N/A                 |   UserImgURL*             |
| **Delete** |  DELETE /blob/user/{user-id/username}        |   N/A                 |   protobuf.empty/{}       |

| Method     |     HTTP Mapping                             |  HTTP Request body    |  HTTP Response body       |
|------------|:--------------------------------------------:|:---------------------:|:-------------------------:|
| **Create** |  POST /blob/author                           |   Multipart           |   AuthorImgURL*           |
| **Get**    |  GET /blob/author/{author-id}                |   N/A                 |   AuthorImgURL*           |
| **Delete** |  DELETE /blob/author/{author-id}             |   N/A                 |   protobuf.empty/{}       |

| Method     |     HTTP Mapping                             |  HTTP Request body    |  HTTP Response body       |
|------------|:--------------------------------------------:|:---------------------:|:-------------------------:|
| **Create** |  POST /blob/media                            |   Multipart           |   MediaURL*               |
| **Get**    |  GET /blob/media/{media-id}                  |   N/A                 |   MediaURL*               |
| **Delete** |  DELETE /blob/media/{media-id}               |   N/A                 |   protobuf.empty/{}       |

