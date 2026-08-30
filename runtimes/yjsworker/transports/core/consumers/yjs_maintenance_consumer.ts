import { type Consumer } from "kafkajs";

import {
  CoreYjsMaintenanceHintTopic,
  YjsCompactionUpdateThreshold,
} from "../../../../../contracts/yjs-worker/v1/yjsworker_contract.js";
import { yjsMaintenanceConfig } from "../../../configs/postgres.js";
import type { YjsPostgresRepository } from "../../../data/postgres/repository.js";
import type { YjsCompactionService } from "../../../services/yjs_compaction_service.js";
import type { YjsProjectionService } from "../../../services/yjs_projection_service.js";
import type { YjsDocumentState } from "../../../types/yjs_document_state.js";
import { createKafkaClient } from "../../../util/kafka.js";
import { Logger } from "../../../util/logger.js";

type EventEnvelope = {
  schemaVersion: string;
  eventType: string;
  aggregateType: string;
  aggregateId: string;
  kafkaKey: string;
  data: unknown;
};

type MaintenanceHint = {
  blockPackId: string;
  documentId: string;
  latestUpdateSequence: number;
  compactedUntilSequence: number;
  projectedUntilSequence: number;
  lastCompactedAt?: string;
  uncompactedUpdateCount: number;
  snapshotBytes: number;
  stateVectorBytes: number;
  reason: string;
};

type MaintenanceResult = {
  success: boolean;
  compactedUntilSequence: number;
  projectedUntilSequence: number;
};

export class YjsMaintenanceConsumer {
  private readonly consumer: Consumer;
  private readonly repository: YjsPostgresRepository;
  private readonly compactionService: YjsCompactionService;
  private readonly projectionService: YjsProjectionService;
  private readonly logger: Logger;
  private readonly pending = new Map<string, MaintenanceHint>();
  private readonly attempts = new Map<string, number>();
  private readonly inFlight = new Set<string>();
  private dispatchPromise: Promise<void> | undefined;
  private started = false;

  constructor(
    repository: YjsPostgresRepository,
    compactionService: YjsCompactionService,
    projectionService: YjsProjectionService,
    logger = new Logger()
  ) {
    const kafka = createKafkaClient();
    this.consumer = kafka.consumer({
      groupId: "notegic-yjs-worker-maintenance-v1",
    });
    this.repository = repository;
    this.compactionService = compactionService;
    this.projectionService = projectionService;
    this.logger = logger;
  }

  async start(): Promise<void> {
    if (this.started) return;

    await this.consumer.connect();
    await this.consumer.subscribe({
      topic: CoreYjsMaintenanceHintTopic,
      fromBeginning: false,
    });
    this.started = true;
    await this.consumer.run({
      eachMessage: async ({ message }) => {
        if (message.value === null) return;

        let event: EventEnvelope;
        try {
          event = JSON.parse(message.value.toString("utf8")) as EventEnvelope;
        } catch {
          return;
        }

        const data = event.data as Partial<MaintenanceHint>;
        if (
          event.schemaVersion !== "v1" ||
          event.eventType !== "YjsMaintenanceHint" ||
          event.aggregateType !== "BlockPack" ||
          event.aggregateId !== event.kafkaKey ||
          typeof data.blockPackId !== "string" ||
          data.blockPackId !== event.aggregateId ||
          typeof data.documentId !== "string" ||
          typeof data.latestUpdateSequence !== "number" ||
          !Number.isSafeInteger(data.latestUpdateSequence) ||
          data.latestUpdateSequence < 0 ||
          typeof data.compactedUntilSequence !== "number" ||
          !Number.isSafeInteger(data.compactedUntilSequence) ||
          data.compactedUntilSequence < 0 ||
          typeof data.projectedUntilSequence !== "number" ||
          !Number.isSafeInteger(data.projectedUntilSequence) ||
          data.projectedUntilSequence < -1
        ) {
          return;
        }

        this.enqueue(data as MaintenanceHint);
        await this.dispatchPending();
      },
    });
  }

  private enqueue(hint: MaintenanceHint): void {
    const existing = this.pending.get(hint.blockPackId);
    if (
      existing === undefined &&
      this.pending.size >= yjsMaintenanceConfig.maximumPendingHints
    ) {
      throw new Error("Yjs maintenance queue is full");
    }
    if (
      existing === undefined ||
      hint.latestUpdateSequence >= existing.latestUpdateSequence
    ) {
      this.pending.set(hint.blockPackId, hint);
    }
  }

  private dispatchPending(): Promise<void> {
    if (this.dispatchPromise !== undefined) return this.dispatchPromise;

    this.dispatchPromise = this.runDispatchPending();
    return this.dispatchPromise;
  }

  private async runDispatchPending(): Promise<void> {
    try {
      while (this.started && this.pending.size > 0) {
        const hints = this.dequeueBatch(
          yjsMaintenanceConfig.maximumDispatchBatch
        );
        for (
          let offset = 0;
          offset < hints.length;
          offset += yjsMaintenanceConfig.maximumDispatchWorkers
        ) {
          const batch = hints.slice(
            offset,
            offset + yjsMaintenanceConfig.maximumDispatchWorkers
          );
          await Promise.all(batch.map(hint => this.process(hint)));
        }
      }
    } finally {
      this.dispatchPromise = undefined;
    }
  }

  private dequeueBatch(limit: number): MaintenanceHint[] {
    const candidates = [...this.pending.values()]
      .filter(hint => !this.inFlight.has(hint.blockPackId))
      .sort((left, right) => {
        const leftScore =
          left.uncompactedUpdateCount * 4 +
          (left.latestUpdateSequence - left.projectedUntilSequence) * 3 +
          (left.lastCompactedAt === undefined ? 100_000 : 0) +
          Math.floor((left.snapshotBytes + left.stateVectorBytes) / 1024);
        const rightScore =
          right.uncompactedUpdateCount * 4 +
          (right.latestUpdateSequence - right.projectedUntilSequence) * 3 +
          (right.lastCompactedAt === undefined ? 100_000 : 0) +
          Math.floor((right.snapshotBytes + right.stateVectorBytes) / 1024);
        return rightScore - leftScore;
      })
      .slice(0, limit);

    for (const hint of candidates) {
      this.pending.delete(hint.blockPackId);
      this.inFlight.add(hint.blockPackId);
    }

    return candidates;
  }

  private async process(hint: MaintenanceHint): Promise<void> {
    const operation =
      hint.uncompactedUpdateCount >= YjsCompactionUpdateThreshold ||
      (hint.compactedUntilSequence < hint.latestUpdateSequence &&
        hint.lastCompactedAt === undefined)
        ? "compact"
        : "project";

    try {
      const result =
        operation === "compact"
          ? await this.compact(hint)
          : await this.project(hint);

      if (
        operation === "compact" &&
        result.success &&
        result.projectedUntilSequence < hint.latestUpdateSequence
      ) {
        this.enqueue({
          ...hint,
          compactedUntilSequence: result.compactedUntilSequence,
          uncompactedUpdateCount: Math.max(
            0,
            hint.latestUpdateSequence - result.compactedUntilSequence
          ),
          projectedUntilSequence: result.projectedUntilSequence,
          lastCompactedAt: new Date().toISOString(),
        });
      }
      this.attempts.delete(hint.blockPackId);
    } catch (error) {
      const attempt = (this.attempts.get(hint.blockPackId) ?? 0) + 1;
      this.attempts.set(hint.blockPackId, attempt);
      if (attempt < yjsMaintenanceConfig.maximumRequestAttempts) {
        setTimeout(() => {
          if (!this.started) return;
          this.enqueue(hint);
          void this.dispatchPending();
        }, 1_000);
      } else {
        this.attempts.delete(hint.blockPackId);
        this.logger.error("Yjs maintenance request failed", {
          blockPackId: hint.blockPackId,
          attempt,
          error: error instanceof Error ? error.message : String(error),
        });
      }
    } finally {
      this.inFlight.delete(hint.blockPackId);
    }
  }

  private async compact(hint: MaintenanceHint): Promise<MaintenanceResult> {
    const loaded = await this.repository.loadCompactable(
      hint.blockPackId,
      hint.latestUpdateSequence
    );
    if (loaded === null) return this.emptyResult(hint);

    const cutoffSequence = Math.min(
      hint.latestUpdateSequence,
      loaded.document.lastUpdateSequence
    );
    if (cutoffSequence <= loaded.document.compactedUntilSequence) {
      return this.emptyResult(hint, loaded.document);
    }

    const compacted = this.compactionService.compact({
      snapshot: loaded.document.snapshot,
      stateVector: loaded.document.stateVector,
      baseCompactedUntilSequence: loaded.document.compactedUntilSequence,
      cutoffSequence,
      updates: loaded.updates,
    });
    const applied = await this.repository.applyCompaction({
      blockPackId: hint.blockPackId,
      ...compacted,
    });

    return {
      success: true,
      compactedUntilSequence: applied.compactedUntilSequence,
      projectedUntilSequence: loaded.document.projectedUntilSequence,
    };
  }

  private async project(hint: MaintenanceHint): Promise<MaintenanceResult> {
    const loaded = await this.repository.loadProjectable(
      hint.blockPackId,
      hint.latestUpdateSequence
    );
    if (loaded === null) return this.emptyResult(hint);

    const targetSequence = Math.min(
      hint.latestUpdateSequence,
      loaded.document.lastUpdateSequence
    );
    if (targetSequence <= loaded.document.projectedUntilSequence) {
      return this.emptyResult(hint, loaded.document);
    }

    const state: YjsDocumentState = {
      snapshot: loaded.document.snapshot,
      stateVector: loaded.document.stateVector,
      lastUpdateSequence: targetSequence,
      compactedUntilSequence: loaded.document.compactedUntilSequence,
      projectedUntilSequence: loaded.document.projectedUntilSequence,
      updates: loaded.updates,
    };
    const projection = this.projectionService.project({
      blockPackId: hint.blockPackId,
      state,
    });
    const applied = await this.repository.applyProjection({
      blockPackId: hint.blockPackId,
      projectedSequence: targetSequence,
      blocks: projection.blocks,
    });

    return {
      success: true,
      compactedUntilSequence: loaded.document.compactedUntilSequence,
      projectedUntilSequence: applied.projectedUntilSequence,
    };
  }

  private emptyResult(
    hint: MaintenanceHint,
    document?: Pick<
      YjsDocumentState,
      "compactedUntilSequence" | "projectedUntilSequence"
    >
  ): MaintenanceResult {
    return {
      success: true,
      compactedUntilSequence:
        document?.compactedUntilSequence ?? hint.compactedUntilSequence,
      projectedUntilSequence:
        document?.projectedUntilSequence ?? hint.projectedUntilSequence,
    };
  }

  async shutdown(): Promise<void> {
    if (!this.started) return;

    this.started = false;
    await this.consumer.disconnect();
    await this.dispatchPromise;
    this.pending.clear();
    this.inFlight.clear();
  }
}
