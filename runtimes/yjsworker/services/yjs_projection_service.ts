import { SpanStatusCode } from "@opentelemetry/api";
import * as Y from "yjs";
import type { Telemetry } from "../telemetry.js";
import {
  createYjsProjectionBatchResult,
  parseYjsProjectionBatchInput,
  type YjsProjectionBatchInput,
  type YjsProjectionBatchResult,
} from "../types/yjs_projection_batch.js";
import type { BlockPackProjector } from "./block_pack_projector.js";

export class YjsProjectionService {
  private readonly blockPackProjector: BlockPackProjector;
  private readonly telemetry: Telemetry;

  constructor(blockPackProjector: BlockPackProjector, telemetry: Telemetry) {
    this.blockPackProjector = blockPackProjector;
    this.telemetry = telemetry;
  }

  project(input: YjsProjectionBatchInput): YjsProjectionBatchResult {
    const startedAt = performance.now();
    const span = this.telemetry.startSpan("projection");
    const document = new Y.Doc();
    try {
      if (input.state.snapshot.length > 0) {
        Y.applyUpdate(document, input.state.snapshot);
      }
      for (const update of input.state.updates) {
        Y.applyUpdate(document, update.payload);
      }

      const result: YjsProjectionBatchResult = {
        blockPackId: input.blockPackId,
        schemaId: "notegic.blocknote",
        schemaVersion: 1,
        projectedSequence: input.state.lastUpdateSequence,
        blocks: this.blockPackProjector.projectYjsDocument(document),
      };
      this.telemetry.recordOperation({
        operation: "projection",
        outcome: "success",
        durationMilliseconds: performance.now() - startedAt,
        payloadBytes: input.state.snapshot.length,
      });

      return result;
    } catch (error) {
      span.recordException(error as Error);
      span.setStatus({ code: SpanStatusCode.ERROR });
      this.telemetry.recordOperation({
        operation: "projection",
        outcome: "error",
        durationMilliseconds: performance.now() - startedAt,
        error,
      });

      throw error;
    } finally {
      document.destroy();
      span.end();
    }
  }

  projectBatch(payload: Buffer): Buffer {
    if (payload.length < 4) {
      throw new Error("invalid yjs projection batch");
    }

    const inputCount = payload.readUInt32BE(0);
    if (inputCount === 0) {
      throw new Error("empty yjs projection batch");
    }

    const outputPayloads: Buffer[] = [];
    const blockPackIds = new Set<string>();
    let offset = 4;
    for (let index = 0; index < inputCount; index += 1) {
      if (payload.length - offset < 4) {
        throw new Error("invalid yjs projection batch");
      }

      const inputLength = payload.readUInt32BE(offset);
      offset += 4;
      if (inputLength > payload.length - offset) {
        throw new Error("invalid yjs projection batch");
      }

      const input = parseYjsProjectionBatchInput(
        payload.subarray(offset, offset + inputLength)
      );
      if (input === null || blockPackIds.has(input.blockPackId)) {
        throw new Error("invalid yjs projection batch");
      }
      blockPackIds.add(input.blockPackId);
      offset += inputLength;

      outputPayloads.push(createYjsProjectionBatchResult(this.project(input)));
    }
    if (offset !== payload.length) {
      throw new Error("invalid yjs projection batch");
    }

    const result = Buffer.alloc(
      4 + outputPayloads.reduce((size, item) => size + 4 + item.length, 0)
    );
    result.writeUInt32BE(outputPayloads.length, 0);

    offset = 4;
    for (const outputPayload of outputPayloads) {
      result.writeUInt32BE(outputPayload.length, offset);
      offset += 4;
      outputPayload.copy(result, offset);
      offset += outputPayload.length;
    }

    return result;
  }
}
