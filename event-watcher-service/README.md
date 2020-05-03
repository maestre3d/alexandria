# Event Watcher Service

The Event Watcher service's main objetive is to **trace every dispatched event** by listening all Alexandria's Event Bus & Message Brokers.
In addition, we shall mention _Alexandria's accepted Message Brokers are Apache Kafka and RabbitMQ._


**We won't be tracking any AWS SNS/SQS, GCP PubSub, NATS or Azure Event Bus Service events.**

Event Watcher service _exposes a tiny API_ using the HTTP protocol and RESTful architectural style.
In a future, we'll be adding a gRPC transport service.

**The Event Watcher service must be kept tiny**, we won't be accepting any complex approaches since this is a helper service for our system architects and devops associates. Our services are not depending on it.

Event Watcher service also integrates with a _web app dashboard GUI_ (made in Angular), it will be available _soon._

***This service must be kept private**, _only for system administrators._