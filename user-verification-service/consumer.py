from confluent_kafka import Consumer, KafkaError

kafka_config = {
    'bootstrap.servers': 'localhost:29092,localhost:29093,localhost:29094',
    'group.id': 'python_consumer',
    'auto.offset.reset': 'earliest'
}

consumer = Consumer(kafka_config)

consumer.subscribe(['create-account'])

print("Listening for messages...")

try:
    while True:
        msg = consumer.poll(timeout=1.0)

        if msg is None:
            continue

        if msg.error():
            if msg.error().code() == KafkaError._PARTITION_EOF:
                continue
            else:
                print(f"Kafka error: {msg.error()}")
                break

        print(f"Received message: {msg.value().decode('utf-8')}")

except KeyboardInterrupt:
    print("Shutting down consumer...")

finally:
    consumer.close()