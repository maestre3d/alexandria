# Alexandria Engineering
We make use of the latest technology to bring a resilient and high-available platform.

Before any specific information, is very important to take a look at our service network to fully understand our needs.

![Alexandria Service Network](https://raw.githubusercontent.com/maestre3d/alexandria/master/docs/Alexandria%20Engineering.png)

In the following section, we show our technology stack selection.

## Development
- **Main programming language**: Go. Created and maintained by Google, Go is our first choice thanks to the easy implementation of complex topics such as concurrency and parallelism, very needed for a high-performance service.
- **Secondary programming language**: Python. With a vast selection of data analysis and machine learning tools, Python makes data-analytic model generation and implementation easier than using any other programming language.
- **Auxiliary programming language**: Javascript. The most popular nowadays, javascript brings simpleness and high productivity thanks to its functional paradigm. Highly used on web and mobile development environments.

## Infrastructure
- **Architecture**: Microservices. Instead of following the typical application architecture (monolithic), we use the microservice architecture to keep our platform high-available, high-scalable and fault-tolerant.
- **Containerization**: Docker and Docker Compose. Thanks to this modern architecture, we use tools like Docker to containerize our builds and keep our services isolated for testing and deployment. In addition, we use Docker Compose to start up our containerized services with just one command.

## Data stores
- **Main data store**: PostgreSQL. Postgres is a robust and high-performing RDBMS which contains useful features that helps to accomplish Alexandria’s daily operations.
- **Main cache data store**: Redis or Memchached.
- **Main graph store**: AWS Neptune or Neo4j.
- **Main document store**: AWS DynamoDB.

## 3rd-party services
- **Message broker/Event bus**: Google Cloud’s Pub/sub, Apache Kafka, RabbitMQ or AWS SQS/SNS.
- **Mailing**: AWS SES.
- **Device push-notifications**: AWS SNS or GCP FCM.
- **Streaming-data analysis**: AWS Kinesis.
- **Batch-data analysis**: On-premise Apache Hadoop or AWS EMR.
- **Identity provider**: AWS Cognito.
- **Secret key management**: Hashicorp Vault or AWS KMS.
- **Static file**: AWS S3/Glacier.
- **Image analysis**: GCP Cloud Vision API.
- **Service discovery/mesh**: Istio, Hashicorp Consul or Netflix Eureka.
- **Serverless functions**: AWS Lambda.
- **Distributed metrics**: AWS Cloudwatch or Prometheus/Grafana.
- **Distributed tracing**: Jaeger, AWS CloudTrail or Zipkin.
- **Distributed logging**: Datadog or Logstash.
- **Querying engine**: Elasticsearch.
- **Content delivery network CDN**: AWS CloudFront.
- **DNS**: AWS Route53.
- **API gateway**: Kubernetes, AWS API Gateway or AWS EC2 instance.
- **Load balancer**: AWS ELB and Netflix Hystrix.
- **Cloud computing**: AWS EC2.

## Deployment
- **Container Orchestration**: Kubernetes -K8s- engine or AWS ECS/Fargate.
- **Testing/Continuous Integration**: Travis CI and GitHub workflows.
- **Continuous Deployment**: Jenkins.

It is **recommended** to use the following PaaS platforms to ensure Alexandria’s performance: 
_Amazon Web Services, Google Cloud Platform or Microsoft Azure_.

Every contributor **must** use any of the defined technologies for each scenario.
