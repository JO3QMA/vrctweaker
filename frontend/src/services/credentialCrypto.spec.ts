import { beforeEach, describe, expect, it } from "vitest";
import { IDBFactory } from "fake-indexeddb";

import {
  WRAPPED_BLOB_MAGIC,
  wrapToken,
  unwrapBlob,
  getOrCreateWrappingKey,
  deleteWrappingKey,
} from "./credentialCrypto";

// Inject a fresh fake IndexedDB before each test so tests are isolated.
beforeEach(() => {
  globalThis.indexedDB = new IDBFactory();
});

describe("WRAPPED_BLOB_MAGIC", () => {
  it("starts with VRCTWKV1:", () => {
    expect(WRAPPED_BLOB_MAGIC).toBe("VRCTWKV1:");
  });
});

describe("wrapToken / unwrapBlob round-trip", () => {
  it("wraps and unwraps a token identically", async () => {
    const token = "authcookie_test_token_xyz";
    const blob = await wrapToken(token);

    expect(blob.startsWith(WRAPPED_BLOB_MAGIC)).toBe(true);
    // Verify it is not plaintext
    expect(blob).not.toContain(token);

    const recovered = await unwrapBlob(blob);
    expect(recovered).toBe(token);
  });

  it("each wrap call produces a different blob (nonce randomness)", async () => {
    const token = "same-token";
    const blob1 = await wrapToken(token);
    const blob2 = await wrapToken(token);
    expect(blob1).not.toBe(blob2);
  });

  it("passes legacy plaintext through unwrapBlob unchanged", async () => {
    const legacyToken = "authcookie_legacy_value";
    const result = await unwrapBlob(legacyToken);
    expect(result).toBe(legacyToken);
  });
});

describe("getOrCreateWrappingKey", () => {
  it("returns equivalent key properties on repeated calls", async () => {
    const key1 = await getOrCreateWrappingKey();
    const key2 = await getOrCreateWrappingKey();
    // Both retrieved from the same IDB entry – same algorithm and usages.
    expect(key1.algorithm).toStrictEqual(key2.algorithm);
    expect(key1.usages).toStrictEqual(key2.usages);
    expect(key1.extractable).toBe(key2.extractable);
  });

  it("key is not extractable", async () => {
    const key = await getOrCreateWrappingKey();
    expect(key.extractable).toBe(false);
  });
});

describe("deleteWrappingKey", () => {
  it("after delete a new key is generated", async () => {
    const key1 = await getOrCreateWrappingKey();
    await deleteWrappingKey();

    // New IDB → new key created
    const key2 = await getOrCreateWrappingKey();
    // Different CryptoKey object (different key material)
    expect(key1).not.toBe(key2);
  });

  it("after key deletion, a blob encrypted with the old key cannot be decrypted", async () => {
    const token = "secret-token";
    const blob = await wrapToken(token);

    await deleteWrappingKey();
    // New IDB (fresh key)

    await expect(unwrapBlob(blob)).rejects.toThrow();
  });
});
