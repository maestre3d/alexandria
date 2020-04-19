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

The following HTTP Proxy Service architecture **should** be used whenever a new HTTP Service is required.

![Alexandria HTTP Proxy Service architecture](https://raw.githubusercontent.com/maestre3d/alexandria/master/docs/Alexandria_http_service.png "Alexandria HTTP Proxy Service architecture")


## Concurrency
We ensure concurrency is being managed in the right way using _oklog’s run_ package referenced in the [OpenCensus’ Go Kit integration guide](https://opencensus.io/integrations/go_kit/).

Every collaborator **must** use oklog run’s group to keep all service/signal handler/cron job & state machine goroutines lifecycles managed correctly.

## Error handling
We expect to keep consistency in error handling; we use multiple packages like _Uber’s multierr_ and _Go’s 1.13 errors_, so we **should** be able to use error wrapping functionality.

In the following section, we show our error handling scenarios.

_**Symbology**_

- RULE(S) - HTTP_CODE/FAULT
    - EXTRA_METHOD -> RETURN_VALUE
        - EXCEPTION_CONTAINED
    - EXCEPTION

_Domain Layer_
- Business rule(s) - 400/user
    - IsValid() -> multierr.combine()
        - InvalidID
        - RequiredField
        - InvalidFieldFormat
        - InvalidFieldRange

_Repository Layer_
- Empty row - 404/not_found
    - Nil Entity
- Already Exists / Unique key - 409/conflict
    - EntityExists
- Infrastructure - 500/internal_server
    - SQL/Docstore_exception(s)

_Use case Layer_
- Parsing - 400/user
    - InvalidID
    - InvalidFieldFormat
    - Required data
    - RequiredField


## Logging
In the following section, we define our logging official sentences.

Every logger **must** write in lowercase. (e.g. http service created)

Every logger **must** specify its layer location separating words with a dot.
_(e.g. service.transport)_

Therefore, every logger **shall** use the specified sentences.
- New instance.- “HANDLER_NAME created”, “LAYER_LOCATION” _(e.g. "media handler created", "service.transport.handler")_
- New service.- “SERVICE_NAME started”, “service.LAYER_LOCATION” _(e.g. "http proxy service started", "service.transport")_

## 3rd-Party Packages
The following specified packages **must** be used for every new service written.
- Gorilla Mux - HTTP Mux/Router
- Go Redis - Redis client
- Google UUID - UUID lib
- Google Wire - Dependency Injection container at compile-time
- Lib pq - Postgres client
- Go Kit - Microservices tool kit
- Oklog Run - Goroutines lifecycle manager
- Spf13 Viper - Configuration manager
- Sony Gobreaker - Circuit breaker library
- Sony Sonyflake - Distributed ID generator, Sony’s implementation of Twitter’s snowflake
- Uber Zap - Logger
- Go Cloud Development Kit (CDK) - Generic kit containing database, pub/sub, document, blob, 
secrets and runtime configuration tools
- Prometheus Client - Prometheus instrumentation client for Go
- Stretchr Testify - Unit Testing lib

_**Optional**_
- NY Times Gizmo - Microservices tool kit
- Spf13 Cobra - CLI tool
- Sirupsen Logrus - Logger
- Uber Multierr - Multi error manager
