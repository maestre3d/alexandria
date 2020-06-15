# Go Guidelines
To keep consistency along all our services, we define our main guidelines so everybody can collaborate.

Take consideration that every line of code written **must** be following our guidelines.

In addition, we’ve taken as reference Uber’s [Go style guide](https://github.com/uber-go/guide/blob/master/style.md).

### Conventions Used in This Document
The requirement level keywords "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" used in this document are to be interpreted as described in [RFC 2119](https://www.ietf.org/rfc/rfc2119.txt).

In this document, such keywords are highlighted using **bold** font.

## Architecture
At the code layer, we should ensure our flexibility and scalability too. We accomplish this with help of a robust architecture.
Either Domain-Driven Design (DDD) by Eric Evans or unclebob’s Clean Architecture **shall** be used to write new services.

### Core concepts
Leveraged from the Hexagonal Architecture, the three main concepts that define our business logic are Entities, Repositories, and Interactors.
- **Entities** are the domain objects (e.g., a Movie or a Shooting Location) — they have no knowledge of where they’re stored (unlike Active Record in Ruby on Rails or the Java Persistence API).
- **Repositories** are the interfaces to getting entities as well as creating and changing them. They keep a list of methods that are used to communicate with data sources and return a single entity or a list of entities. (e.g. UserRepository)
- **Interactors** are classes that orchestrate and perform domain actions — think of Service Objects or Use Case Objects. They implement complex business rules and validation logic specific to a domain action (e.g., onboarding a production)


With these three main types of objects, we are able to define business logic without any knowledge or care where the data is kept and how business logic is triggered. Outside of the business logic are the Data Sources and the Transport Layer:
- **Data Sources** are adapters to different storage implementations.
A data source might be an adapter to a SQL database (an Active Record class in Rails or JPA in Java), an elastic search adapter, REST API, or even an adapter to something simple such as a CSV file or a Hash. A data source implements methods defined on the repository and stores the implementation of fetching and pushing the data.
- **Transport Layer** can trigger an interactor to perform business logic. We treat it as an input for our system. The most common transport layer for microservices is the HTTP **API Layer** and a set of controllers that handle requests. By having business logic extracted into interactors, we are not coupled to a particular transport layer or controller implementation. Interactors can be triggered not only by a controller, but also by an event, a cron job, or from the command line.

![Clean architecture](https://miro.medium.com/max/1400/1*NfFzI7Z-E3ypn8ahESbDzw.png "Alexandria HTTP Proxy Service architecture")
_From Netflix Engineering, [click here](https://netflixtechblog.com/ready-for-changes-with-hexagonal-architecture-b315ec967749)._

Therefore, we use our own API architecture implementing a Proxy pattern to hide our transport services implementation (like HTTP).

The following application architecture **should** be used whenever a new service is created.

![Alexandria HTTP Proxy Service architecture](https://raw.githubusercontent.com/maestre3d/alexandria/master/docs/Alexandria%20Service.png "Alexandria HTTP Proxy Service architecture")


## Concurrency
We ensure concurrency is being managed in the right way using _oklog’s run_ package referenced in the [OpenCensus’ Go Kit integration guide](https://opencensus.io/integrations/go_kit/).

Every collaborator **must** use oklog run’s group to keep all service/signal handler/cron job & state machine goroutines lifecycles managed correctly.

Every main function **must** create a new context from background with a cancellation function. Eventually, this root context **must** be injected inside the dependecy injection package. Therefore, the cancellation function **must** be triggered inside the system's exit listener whenever the listener it's triggered, so the root context gets cancelled properly.

Every infrastructure call from a use case function **must** create it's own context.

## Error handling
We expect to keep consistency in error handling; we use multiple packages like _Uber’s multierr_ and _Go’s 1.13 errors_, so we **should** be able to use error wrapping functionality.

In the following section, we show our error handling scenarios by layer.


**Domain**: _Business rule(s) validations_

| Type                   |     Description                                            |  HTTP Status Code     |  Return value     |
|------------------------|------------------------------------------------------------|:---------------------:|:---------------------:|
| **InvalidID**          |  Invalid identifier                                        |   400                 |   Exception                 | 
| **RequiredField**      |  Missing required request field _x_                        |   400                 |   Exception                 |
| **InvalidFieldFormat** |  Request field _x_ has an invalid format, expect _value_   |   400                 |   Exception                 |
| **InvalidFieldRange**  |  Request field _x_ is out of range `[x, y)`                |   400                 |   Exception                 |



**Repository**: _Data source(s) validations_

| Type                    |     Description                                                 |  HTTP Status Code   |  Return value     |
|-------------------------|-----------------------------------------------------------------|:-------------------:|:---------------------:|
| **EmptyRow**            |  Resource(s) not found                  |   404                 |   Null/Nil          |
| **Infrastructure**      |  SQL/Docstore/API internal error        |   500                 |   Exception         |
| **AlreadyExists**       |  Resource already exists                |   409                 |   Exception         |


**Interactor**: _Domain's cases validation_

| Type                   |     Description                                            |  HTTP Status Code     |  Return value     |
|------------------------|------------------------------------------------------------|:---------------------:|:---------------------:|
| **InvalidID**          |  Invalid identifier                                        |   400                 |   Exception                 | 
| **RequiredField**      |  Missing required request field _x_                        |   400                 |   Exception                 |
| **InvalidFieldFormat** |  Request field _x_ has an invalid format, expect _value_   |   400                 |   Exception                 |
| **InvalidFieldRange**  |  Request field _x_ is out of range `[x, y)`                |   400                 |   Exception                 |
| **AlreadyExists**      |  Resource already exists                                   |   409                 |   Exception                 |
| **EmptyBody**          |  Request body is empty                                     |   400                 |   Exception                 |



## Logging
In the following section, we define our logging official sentences.

Every logger **must** write in lowercase. (e.g. http service created)

Every logger **must** specify its layer location separating words with a dot.
_(e.g. service.transport)_

Therefore, every logger **shall** use the specified sentences.
- New instance.- “HANDLER_NAME created”, “LAYER_LOCATION” _(e.g. "media handler created", "service.transport.handler")_
- New service.- “SERVICE_NAME started”, “service.LAYER_LOCATION” _(e.g. "http proxy service started", "service.transport")_

## Runtime Configuration
In the following section, we define our runtine configuration guideline.

Every configuration **must** define default values inside code.

Every configuration **must** have an _"alexandria-config.yaml"_ file containing required keys, it must be stored in the following locations:
- _$HOME/.alexandria/_
- _./config/_
- _/etc/alexandria/_
- _._

Every configuration system **must** fetch secrets from _AWS KMS_ or similar. If not available, read configuration from _"alexandria-config.yaml"_ file.

**Configuration file example**
```yaml
alexandria:
  info:
    service: "media"
    version: 1.0.0
  persistence:
    dbms:
      url: "postgres://postgres:root@postgres:5432/alexandria_media?sslmode=disable"
      driver: "postgres"
      user: "postgres"
      password: "root"
      host: "postgres"
      port: 5432
      database: "alexandria_media"
    mem:
      network: ""
      host: "redis"
      port: 6379
      password: ""
      database: 0
  service:
    transport:
      http:
        host: "0.0.0.0"
        port: 8080
      rpc:
        host: "0.0.0.0"
        port: 31337
```

## Versioning
For every single release, all software deployed **must** use [Semantic Versioning](https://semver.org/) guidelines to keep consistency and inform every single developer the best way possible.

## 3rd-Party Packages
The following specified packages **must** be used for every new service written.
- Alexandria-OSS Core - Main core for our services
- Gorilla Mux - HTTP Mux/Router
- Go Redis - Redis client
- Google UUID - UUID lib
- Matoous NanoID - Nano ID implementation in Go
- Google Wire - Dependency Injection container at compile-time
- Lib pq - Postgres client
- Go Kit - Microservices tool kit
- Oklog Run - Goroutines lifecycle manager
- Spf13 Viper - Configuration manager
- Sony Gobreaker - Circuit breaker library
- Sony Sonyflake - Distributed ID generator, Sony’s implementation of Twitter’s snowflake
- Uber Zap - Logger
- Google Go Cloud Development Kit (CDK) - Generic kit containing database, pub/sub, document, blob, 
secrets and runtime configuration tools
- Prometheus Client - Prometheus instrumentation client for Go
- Stretchr Testify - Unit Testing lib
- Go Playground validator - Struct/Model validations

_**Optional**_
- NY Times Gizmo - Microservices tool kit
- Spf13 Cobra - CLI tool
- Sirupsen Logrus - Logger
- Uber Multierr - Multi error manager
- Jaeger Tracer - OpenCensus/OpenTracing consumer
- Elasticsearch - Elasticsearch client for Go
