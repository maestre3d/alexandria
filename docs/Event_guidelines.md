# Event guidelines

In Alexandria, we make use of the event-driven architectural style to accomplish the "loosely-coupled" concept mentioned 
in microservices terminologies. We communicate our services through different kind of events asynchronously.

Besides from the benfits convetional way to communicate services (synchronous) gives, an event-driven environment gives us 
the flexibility to scale easily and additionally, it gives us the resiliency we need to keep our services healthy 
and available. Also, it avoids the need of Service Discovery/Service Mesh mechanisms in every service we develop.

Nevertheless, we fall inside in a lot of complexity thanks to this new architectural style. Thus, we have written 
this guidelines so any contributor will be able to fully understand our internal event architecture.

In the following section, we define our guidelines for event consuming and producing along with some terminology  
that every contributor must follow.

## Events
There is two event types:
- **Integration**: This represents a _transaction_, mostly SAGAs. They **must** contain an root span ID for tracing,
transaction ID and operation type if applicable. Also, every event must consider a rollback operation for 
every single commit operation. _They are immutable._
- **Domain**: This represents an _event with side-effects_, does not need any transaction. This is used
to propagate effects in our distributed environment. Any service _could_ be listening to it. _They are mutable._

Besides of these types, you _should_ follow the next rules:
- Every event **must** have inside its metadata the fields: 
    - ID: Event unique identifier.
    - Service: Service who sends the event in upper case (could be using the server config file).
    - Event type: Integration or Domain event.
    - Priority: High, Mid or Low.
    - Provider: Message broker used.
    - Dispatch Time: Unix timestamp.
- Every event _should_ have a body, it _could_ contain entities, aggregates or just entity's ID.

## Transactions
In every event-driven architecture, using transactions is a **must**-have to keep data up to date.
In the following section, we will discuss about transactions in a deep way.

This represents an integration event.
Transactions **must** use the following body:
  - ID: Transaction ID
  - Root ID: Aggregation/Entity ID.
  - Span ID: OpenTracing/OpenCensus root span ID.
  - Trace ID: OpenTracing/OpenCensus trace ID.
  - Operation: Kind of operation to perform. (e.g. FOO_CREATED, FOO_UPDATED)
  - Backup: Aggregate/Entity's backup for update-like operations. This _could_ be skipped.
  
If a transaction fails, you _should_ replace the event body with a generic response like this:
  - Code: HTTP status-like code.
  - Message: Event's extra information about transaction (useful for logging).

### Best Practices
- If you need to verify external entities from your service using events, you **must** send an string array/slice with all the IDs (reference/foreign keys) you want to verify as event's JSON body/content.
- If your service verifies unique identifiers, you **must** add a Verify() function inside you SAGA interactor and your SAGA event bus **must** have a Verified() and Failed() functions.
- The SAGA event bus **must** receive the correct event context to respond successfully.
- Verified() and Failed() event functions **must** use the event's entity "service" field to attach the name to the domain's event name (e.g. Service Name = SERVICEFOO, Domain's Event Name = AUTHOR_VERIFIED, Result = SERVICEFOO_AUTHOR_VERIFIED), this is done to avoid event overlapping with other services.

## Producer
This represents the publisher of a topic.
- Every event **must** be written in _uppercase_ and using a verb _written in simple past_. (e.g.  AUTHOR_DELETED)
- Every event **must** use the respective event type (integration or domain)
- At codebase, every producer **must** be triggered at the interaction layer (use case domain) and executed at the 
infrastructure implementation.

## Consumer
This represents the subscription or listener of a topic.
-  At codebase, every consumer **must** have the _preffix 'on'_ and **must** contain a verb _written in simple past_.
_(e.g. 'onUserDeleted')_
-  Every consumer **must** have a _queue system attached_ to the topic to accomplish resiliency, so if the service 
becomes unavailable, it will pull all the pending messages from the message broker whenever it gets back online.
In addition, a queue system also gives us a new way to handle messages inside a service cluster. Thus, data consistency 
is being respected.
In Apache Kafka, these queues are represented by _consumer groups_. Likewise, in AWS we use SQS to accomplish the same 
goal.
- Every service cluster **must** have it's own _consumer group/queue_ so messages will be acknowledged consistently.
- At codebase, every consumer **must** be listening inside the transport layer.

## Tracing
Even though we are using asynchronous communication, we **must** still trace our operations to enable observability.
We recommend using either OpenCensus or OpenTracing APIs to do this, along with a tracer like Jaeger or Zipkin.
This is mostly recommended for every integration event (transaction).

## Context Propagation
Since almost any service will perform transactions, context propagation is a nice way to handle transaction chaining.
We recommend following the OpenCensus/OpenTracing approaches.

To keep transactions and events intact, you **must** use context propagation.

An event context **must** have the following fields:
- Transaction entity: The root transaction entity, this is immutable.
- Event entity: The parent's event entity, this is mutable.

## Best Practices
- Do not acknowledge messages if something happened at your infrastructure layer 
(database connection closed, error while openning topic, etc).
- Acknowledge messages if a client-related error happened (data parsing, malformed event context, etc) to avoid processing loops.
- Send the complete entity for every domain event sent.
- Write opposite use case functions in case of a rollback. (e.g. create-hard_delete)
  - Create a createRaw() function in every repository to restore from a hard delete operation.
- Segregate organic business cases (normal use case) from SAGA use cases, create a new interactor class/struct for SAGA-only operations.
  - Write Done() and Failed() functions for SAGA commit/rollback transactions. (Keep your main functions organic)
- For producer calling, add a new function in your Organic/SAGA interactor and trigger the producer from the event bus at the infrastructure layer.
- For replace operations, attach the old entity in the event's metadata as a 'backup' field in case of a rollback.
- For create operation's rollback, execute hard delete.
- Log whenever a new event was produced.
- Inject metrics from Prometheus or OpenMetrics.
- Use distributed tracing instrumentation.
- Create a new context for each consumer. _*If applicable_
- If event producing fails and you performed a write operation, rollback immediately using straight repository's methods.
- For every entity that requires transactions, add a new field called 'status' and create an enum with all the 
states available, so the client's could know the current status.
- Use short polling to receive updates in _"real time"_ on the client (React/Angular/Android/iOS, etc),
avoid websockets or SSE if possible.
- There is three possible states for recently created entity's, _'proccessing', 'done' and 'failed'_,
and they _could_ be represented like this:
  - Proccessing: Client calls to API with /GET, receives the created resource/entity but with status different than done.
  - Done: Client calls to API with /GET, receives the created resource/entity with status equal to done.
  - Failed: Client calls to API with /GET, receives a 404/NotFound error.
- Wrap topics with resiliency patterns like circuit breaker(hystrix, gobreaker, etc) to avoid common communication errors.
- Propagate the event context like OpenTracing/OpenCensus to keep the transaction and event entities intact.

