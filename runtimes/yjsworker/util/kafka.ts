import { Kafka, logLevel } from "kafkajs";

export function createKafkaClient(): Kafka {
  const brokers = (process.env.KAFKA_BROKERS ?? "127.0.0.1:9094")
    .split(",")
    .map(broker => broker.trim())
    .filter(Boolean);

  return new Kafka({
    clientId: process.env.KAFKA_CLIENT_ID ?? "notegic-yjs-worker",
    brokers,
    logLevel: logLevel.NOTHING,
  });
}
