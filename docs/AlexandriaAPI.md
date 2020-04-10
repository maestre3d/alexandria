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


“In resource-oriented APIs, resources are named entities, and resource names are their identifiers. Each resource must have its own unique resource name. A collection is a special kind of resource that contains a list of sub-resources of identical type. The resource ID for a collection is called collection ID.” (Google, 2020)

In the following list, we define our API’s resources and collection IDs.

- **/users** -> List, Create
    - /{user-id} – {username} -> Get, Update, Delete
- **/authors** -> List, Create
    - /{author-id} -> Get, Update, Delete
- **/blobs**
    - **/users** -> Create
        - /{username} – {user-id} -> Get, Delete
    - **/authors** -> Create
        - /{author-id} -> Get, Delete
    - **/media** -> Create
        - /{media-id} -> Get, Delete
- **/media** -> List, Create
    - /{media-id} -> Get, Update, Delete
- **/categories** -> List, Create
    - /{category-name} – {category-id} -> Get, Update, Delete
        - **/media** -> Create
            - /{media-id} -> Get, Delete
        - **/authors** -> Create
            - /{author-id} -> Get, Delete
