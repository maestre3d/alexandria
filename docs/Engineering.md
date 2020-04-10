# Alexandria Engineering
We make use of the latest technology to bring a resilient and high-available platform.

In the following section, we show our technology stack selection.

## Development
- **Main programming language**: Go. Created and maintained by Google, Go is our first choice thanks to the easy implementation of complex topics such as concurrency and parallelism, very needed for a high-performance service.
- **Secondary programming language**: Python. With a vast selection of data analysis and machine learning tools, Python makes data-analytic model generation and implementation easier than using any other programming language.

## Infrastructure
- **Architecture**: Microservices. Instead of following the typical application architecture (monolithic), we use the microservice architecture to keep our platform high-available, high-scalable and fault-tolerant.
- **Containerization**: Docker and Docker Compose. Thanks to this modern architecture, we use tools like Docker to containerize our builds and keep our services isolated for testing and deployment. In addition, we use Docker Compose to start up our containerized services with just one command-line.

## Data stores
- **Main data store**: PostgreSQL. Postgres is a robust and high-performing RDBMS which contains useful features that helps to accomplish Alexandria’s daily operations.
- **Main cache data store**: Redis.

## 3rd-party services
- **Microservice communication**: Google Cloud’s Pub/sub, Apache Kafka, RabbitMQ, NATS or AWS SQS/SNS.
- **Mailing**: AWS SES.
- **Device push-notifications**: AWS SNS or GCP FCM.
- **Streaming-data analysis**: AWS Kinesis.
- **Batch-data analysis**: On-premise Apache Hadoop or AWS EMR.

## Deployment
- **Container Orchestration**: Kubernetes -K8s- engine.
- **Testing/Continuous Integration**: Travis CI and GitHub workflows.
- **Continuous Deployment**: Jenkins.

It is **recommended** to use the following PaaS platforms to ensure Alexandria’s performance: 
_Amazon Web Services, Google Cloud Platform or Microsoft Azure_.

Every contributor **must** use any of the defined technologies for each scenario.