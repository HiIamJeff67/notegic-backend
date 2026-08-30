import { randomUUID } from "node:crypto";
import { createRequire } from "node:module";
import { type Consumer, Kafka, type Producer } from "kafkajs";

import {
  CoreYjsWorkerReplyTopic,
  YjsWorkerCoreCommandTopic,
} from "../../../../../contracts/yjs-worker/v1/yjsworker_contract.js";
import { createKafkaClient } from "../../../util/kafka.js";
import { Logger } from "../../../util/logger.js";

export {
  CoreYjsWorkerReplyTopic,
  YjsWorkerCoreCommandTopic,
} from "../../../../../contracts/yjs-worker/v1/yjsworker_contract.js";

type EventEnvelope = {
  schemaVersion: string;
  eventId: string;
  eventType: string;
  aggregateType: string;
  aggregateId: string;
  kafkaKey: string;
  occurredAt: string;
  correlationId: string;
  causationId?: string;
  trace: {
    traceParent?: string;
    traceState?: string;
  };
  data: unknown;
};

type CommandEnvelope<D> = {
  schemaVersion: string;
  commandId: string;
  commandType: string;
  blockPackId: string;
  correlationId: string;
  trace: {
    traceParent?: string;
    traceState?: string;
  };
  producer: string;
  occurredAt: string;
  data: D;
};

type ReplyEnvelope<D> = {
  schemaVersion: string;
  commandId: string;
  commandType: string;
  blockPackId: string;
  correlationId: string;
  producer: string;
  respondedAt: string;
  data: D;
  error?: {
    code: string;
    message: string;
    retryable: boolean;
  };
};

type PendingReply = {
  resolve: (data: unknown) => void;
  reject: (error: Error) => void;
  timeout: NodeJS.Timeout;
};

export class CoreCommandError extends Error {
  readonly code: string | null;
  readonly retryable: boolean | null;
  readonly isTimeout: boolean;
  readonly commandId: string | null;
  readonly commandType: string;

  constructor(
    message: string,
    options: {
      code?: string | null;
      retryable?: boolean | null;
      isTimeout?: boolean;
      commandId?: string | null;
      commandType: string;
    }
  ) {
    super(message);
    this.name = "CoreCommandError";
    this.code = options.code ?? null;
    this.retryable = options.retryable ?? null;
    this.isTimeout = options.isTimeout ?? false;
    this.commandId = options.commandId ?? null;
    this.commandType = options.commandType;
  }
}

type CoreCommand = {
  commandId: string;
  event: EventEnvelope;
};

const require = createRequire(import.meta.url);
const kafkajs = require("kafkajs") as typeof import("kafkajs");

kafkajs.CompressionCodecs[
  kafkajs.CompressionTypes.Snappy
] = require("kafkajs-snappy");

export class CoreCommandProducer {
  private readonly producer: Producer;
  private started = false;

  constructor(kafka: Kafka) {
    this.producer = kafka.producer();
  }

  async start(): Promise<void> {
    if (this.started) return;

    await this.producer.connect();
    this.started = true;
  }

  create<D>(commandType: string, blockPackId: string, data: D): CoreCommand {
    const commandId = randomUUID();
    const correlationId = randomUUID();
    const occurredAt = new Date().toISOString();
    const command: CommandEnvelope<D> = {
      schemaVersion: "v1",
      commandId,
      commandType,
      blockPackId,
      correlationId,
      trace: {},
      producer: "yjs-worker",
      occurredAt,
      data,
    };

    return {
      commandId,
      event: {
        schemaVersion: "v1",
        eventId: commandId,
        eventType: "YjsWorkerCommand",
        aggregateType: "BlockPack",
        aggregateId: blockPackId,
        kafkaKey: blockPackId,
        occurredAt,
        correlationId,
        causationId: commandId,
        trace: {},
        data: command,
      },
    };
  }

  async produce(command: CoreCommand): Promise<void> {
    await this.start();
    await this.producer.send({
      topic: YjsWorkerCoreCommandTopic,
      messages: [
        { key: command.event.kafkaKey, value: JSON.stringify(command.event) },
      ],
    });
  }

  async shutdown(): Promise<void> {
    if (!this.started) return;

    await this.producer.disconnect();
    this.started = false;
  }
}

export class CoreReplyConsumer {
  private readonly consumer: Consumer;
  private readonly logger: Logger;
  private readonly groupId: string;
  private readonly pendingReplies = new Map<string, PendingReply>();
  private startPromise: Promise<void> | null = null;
  private started = false;

  constructor(kafka: Kafka, logger = new Logger()) {
    const instanceId = process.env.HOSTNAME ?? randomUUID();
    this.groupId = `notegic-yjsworker-replies-${instanceId}`;
    this.logger = logger;
    this.consumer = kafka.consumer({ groupId: this.groupId });
    this.consumer.on(this.consumer.events.CONNECT, () => {
      this.logger.debug("[YjsWorker] Core reply consumer connected", {
        groupId: this.groupId,
        topic: CoreYjsWorkerReplyTopic,
      });
    });
    this.consumer.on(this.consumer.events.GROUP_JOIN, event => {
      this.logger.debug("[YjsWorker] Core reply consumer joined group", {
        groupId: this.groupId,
        memberId: event.payload.memberId,
        topics: Object.keys(event.payload.memberAssignment),
      });
    });
    this.consumer.on(this.consumer.events.DISCONNECT, () => {
      this.logger.warn("[YjsWorker] Core reply consumer disconnected", {
        groupId: this.groupId,
      });
    });
    this.consumer.on(this.consumer.events.CRASH, event => {
      this.logger.error("[YjsWorker] Core reply consumer crashed", {
        groupId: this.groupId,
        restart: event.payload.restart,
        error: event.payload.error.message,
        stack: event.payload.error.stack,
      });
    });
  }

  async start(): Promise<void> {
    if (this.started) return;

    if (this.startPromise !== null) {
      await this.startPromise;

      return;
    }

    this.startPromise = this.startConsumer();
    try {
      await this.startPromise;
    } finally {
      this.startPromise = null;
    }
  }

  private async startConsumer(): Promise<void> {
    this.logger.debug("[YjsWorker] Starting Core reply consumer", {
      groupId: this.groupId,
      topic: CoreYjsWorkerReplyTopic,
    });
    await this.consumer.connect();
    await this.consumer.subscribe({
      topic: CoreYjsWorkerReplyTopic,
      fromBeginning: false,
    });

    const groupJoined = new Promise<void>(resolve => {
      const removeListener = this.consumer.on(
        this.consumer.events.GROUP_JOIN,
        () => {
          removeListener();
          resolve();
        }
      );
    });

    await this.consumer.run({
      eachMessage: async ({ message }) => {
        if (message.value === null) return;

        let event: EventEnvelope;
        try {
          event = JSON.parse(message.value.toString("utf8")) as EventEnvelope;
        } catch {
          return;
        }
        const reply = event.data as ReplyEnvelope<unknown>;
        if (
          event.schemaVersion !== "v1" ||
          event.eventType !== "YjsWorkerCommandCompleted" ||
          reply.schemaVersion !== "v1" ||
          typeof reply.commandId !== "string"
        ) {
          return;
        }

        const pendingReply = this.pendingReplies.get(reply.commandId);
        if (pendingReply === undefined) {
          this.logger.debug("[YjsWorker] Ignored unmatched Core reply", {
            commandId: reply.commandId,
            commandType: reply.commandType,
          });
          return;
        }

        this.pendingReplies.delete(reply.commandId);
        clearTimeout(pendingReply.timeout);
        this.logger.debug("[YjsWorker] Received Core reply", {
          commandId: reply.commandId,
          commandType: reply.commandType,
          hasError: reply.error !== undefined,
        });
        if (reply.error !== undefined) {
          pendingReply.reject(
            new CoreCommandError(
              `${reply.error.code}: ${reply.error.message}`,
              {
                code: reply.error.code,
                retryable: reply.error.retryable,
                commandId: reply.commandId,
                commandType: reply.commandType,
              }
            )
          );

          return;
        }
        pendingReply.resolve(reply.data);
      },
    });

    await groupJoined;
    this.started = true;
  }

  waitForReply<R>(commandId: string, commandType: string): Promise<R> {
    return new Promise<R>((resolve, reject) => {
      const timeout = setTimeout(
        () => {
          this.pendingReplies.delete(commandId);
          reject(
            new CoreCommandError(`YjsWorker command ${commandType} timed out`, {
              code: "Timeout",
              retryable: true,
              isTimeout: true,
              commandId,
              commandType,
            })
          );
        },
        Number(process.env.YJS_WORKER_COMMAND_TIMEOUT_MILLISECONDS ?? 10_000)
      );
      this.pendingReplies.set(commandId, {
        resolve: data => resolve(data as R),
        reject,
        timeout,
      });
    });
  }

  reject(commandId: string, error: unknown): void {
    const pendingReply = this.pendingReplies.get(commandId);
    if (pendingReply === undefined) return;

    this.pendingReplies.delete(commandId);
    clearTimeout(pendingReply.timeout);
    pendingReply.reject(
      error instanceof Error ? error : new Error(String(error))
    );
  }

  async shutdown(): Promise<void> {
    for (const pendingReply of this.pendingReplies.values()) {
      clearTimeout(pendingReply.timeout);
      pendingReply.reject(new Error("YjsWorker is shutting down"));
    }
    this.pendingReplies.clear();
    if (!this.started) return;

    await this.consumer.disconnect();
    this.started = false;
  }
}

export class CoreCommandDispatcher {
  private readonly producer: CoreCommandProducer;
  private readonly replyConsumer: CoreReplyConsumer;

  constructor(logger = new Logger()) {
    const kafka = createKafkaClient();
    this.producer = new CoreCommandProducer(kafka);
    this.replyConsumer = new CoreReplyConsumer(kafka, logger);
  }

  async dispatch<D, R>(
    commandType: string,
    blockPackId: string,
    data: D
  ): Promise<R> {
    await this.producer.start();
    await this.replyConsumer.start();

    const command = this.producer.create(commandType, blockPackId, data);
    const reply = this.replyConsumer.waitForReply<R>(
      command.commandId,
      commandType
    );
    try {
      await this.producer.produce(command);
    } catch (error) {
      this.replyConsumer.reject(command.commandId, error);
    }

    return reply;
  }

  async dispatchAsync<D, R>(
    commandType: string,
    blockPackId: string,
    data: D
  ): Promise<{ commandId: string; reply: Promise<R> }> {
    await this.producer.start();
    await this.replyConsumer.start();

    const command = this.producer.create(commandType, blockPackId, data);
    const reply = this.replyConsumer.waitForReply<R>(
      command.commandId,
      commandType
    );
    try {
      await this.producer.produce(command);
    } catch (error) {
      this.replyConsumer.reject(command.commandId, error);
      throw error;
    }

    return { commandId: command.commandId, reply };
  }

  async shutdown(): Promise<void> {
    await this.replyConsumer.shutdown();
    await this.producer.shutdown();
  }
}
