1. The *controller* publishes each instruction as a record to a Kafka topic.
2. Each *worker* reads each record from Kafka.
3. Each *worker* parses the JSON text published to Kafka and adds their light control instructions to their queue.

Workers are, in the background, executing their queue.