# Go Guidelines
To keep consistency along all our services, we define our main guidelines so anybody can collaborate.

Take consideration that every line of code written **must** be following our guidelines.

In addition, we’ve taken as reference Uber’s [Go style guide](https://github.com/uber-go/guide/blob/master/style.md).

### Conventions Used in This Document
The requirement level keywords "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" used in this document are to be interpreted as described in [RFC 2119](https://www.ietf.org/rfc/rfc2119.txt).

In this document, such keywords are highlighted using **bold** font.

## Architecture
At the code layer, we should ensure our flexibility and scalability too. We accomplish this with help of a robust architecture.
Either Domain-Driven Design (DDD) by Eric Evans or unclebob’s Clean Architecture **shall** be used to write new services.

Therefore, we use our own API architecture implementing a Proxy pattern to hide our delivery services implementation (like HTTP).

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
    - EntityNotFound
    - EntitiesNotFound
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
_(e.g. service.delivery)_

Therefore, every logger **shall** use the specified sentences.
- New instance.- “HANDLER_NAME created”, “LAYER_LOCATION” _(e.g. "media handler created", "service.delivery.handler")_
- New service.- “SERVICE_NAME started”, “service.LAYER_LOCATION” _(e.g. "http proxy service started", "service.delivery")_

## 3rd-Party Packages
The following specified packages **must** be used for every new service written.
- Gin - HTTP Mux/Router
- Go Redis - Redis client
- Google UUID - UUID lib
- Google Wire - Dependency Injection container generator
- Lib PQ - Postgres client
- Oklog Run - Goroutines lifecycle manager
- Sony Gobreaker - Circuit breaker library
- Sony Sonyflake - Distributed ID generator, Sony’s implementation of Twitter’s snowflake
- Uber Multierr - Multi error manager
- Uber Zap - Logger
- Go Cloud Development Kit (CDK) - Generic kit containing database, pub/sub, document, blob, 
secrets and runtime configuration tools

_**Optional**_
- Go Kit - Microservices tool kit
- NY Times Gizmo - Microservices tool kit
