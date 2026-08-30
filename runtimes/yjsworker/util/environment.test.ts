import assert from "node:assert/strict";
import test from "node:test";

import { port, positiveInteger, required } from "./environment.js";

function withEnvironment(
  name: string,
  value: string | undefined,
  operation: () => void
): void {
  const previousValue = process.env[name];
  if (value === undefined) {
    delete process.env[name];
  } else {
    process.env[name] = value;
  }

  try {
    operation();
  } finally {
    if (previousValue === undefined) {
      delete process.env[name];
    } else {
      process.env[name] = previousValue;
    }
  }
}

test("reads required environment variables after trimming whitespace", () => {
  withEnvironment("YJS_TEST_REQUIRED", "  value  ", () => {
    assert.equal(required("YJS_TEST_REQUIRED"), "value");
  });
});

test("uses a fallback for an unset port and validates the range", () => {
  withEnvironment("YJS_TEST_PORT", undefined, () => {
    assert.equal(port("YJS_TEST_PORT", 8787), 8787);
  });

  withEnvironment("YJS_TEST_PORT", "", () => {
    assert.equal(port("YJS_TEST_PORT", 8787), 8787);
  });

  withEnvironment("YJS_TEST_PORT", "5432", () => {
    assert.equal(port("YJS_TEST_PORT"), 5432);
  });

  withEnvironment("YJS_TEST_PORT", "65536", () => {
    assert.throws(
      () => port("YJS_TEST_PORT"),
      /YJS_TEST_PORT must be an integer between 1 and 65535/
    );
  });
});

test("validates positive integer environment variables", () => {
  withEnvironment("YJS_TEST_INTEGER", "12", () => {
    assert.equal(positiveInteger("YJS_TEST_INTEGER", 10), 12);
  });

  withEnvironment("YJS_TEST_INTEGER", "0", () => {
    assert.throws(
      () => positiveInteger("YJS_TEST_INTEGER", 10),
      /YJS_TEST_INTEGER must be a positive integer/
    );
  });
});
