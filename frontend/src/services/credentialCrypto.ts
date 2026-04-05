/**
 * credentialCrypto – AES-GCM wrapping/unwrapping for VRChat auth tokens.
 *
 * Security notes:
 *  - The CryptoKey is generated with `extractable: false` so `exportKey` is blocked.
 *  - Nonce is always generated via `crypto.getRandomValues` (12 random bytes per NIST SP 800-38D).
 *    Counter, fixed, or time-derived nonces are PROHIBITED to prevent GCM key-stream exposure.
 *  - Wrapped blobs must start with WRAPPED_BLOB_MAGIC; this distinguishes them from legacy
 *    plaintext tokens stored before Phase-B migration.
 */

const DB_NAME = "vrctwk-credentials";
const STORE_NAME = "keys";
const KEY_ID = "auth-key";
const KEY_ALGO: AesKeyGenParams = { name: "AES-GCM", length: 256 };

/** Magic prefix that Go uses to distinguish wrapped blobs from legacy plaintext. */
export const WRAPPED_BLOB_MAGIC = "VRCTWKV1:";

/** AES-GCM nonce length in bytes (NIST recommended 96-bit IV). */
const NONCE_LENGTH = 12;

// ---------------------------------------------------------------------------
// IndexedDB key persistence
// ---------------------------------------------------------------------------

function openKeyDb(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, 1);
    req.onupgradeneeded = () => {
      req.result.createObjectStore(STORE_NAME);
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

/** Opens the DB, runs fn, then closes the connection after all work completes. */
async function withKeyDb<T>(fn: (db: IDBDatabase) => Promise<T>): Promise<T> {
  const db = await openKeyDb();
  try {
    return await fn(db);
  } finally {
    db.close();
  }
}

function idbGet(db: IDBDatabase, id: string): Promise<CryptoKey | undefined> {
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, "readonly");
    const req = tx.objectStore(STORE_NAME).get(id);
    let value: CryptoKey | undefined;
    req.onsuccess = () => {
      value = req.result as CryptoKey | undefined;
    };
    req.onerror = () => reject(req.error);
    tx.onerror = () =>
      reject(tx.error ?? new Error("IndexedDB transaction failed"));
    tx.oncomplete = () => resolve(value);
  });
}

function idbPut(db: IDBDatabase, id: string, key: CryptoKey): Promise<void> {
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, "readwrite");
    const req = tx.objectStore(STORE_NAME).put(key, id);
    req.onerror = () => reject(req.error);
    tx.onerror = () =>
      reject(tx.error ?? new Error("IndexedDB transaction failed"));
    tx.oncomplete = () => resolve();
  });
}

/**
 * Returns the persisted AES-GCM wrapping key, creating a new one if absent.
 * The key is `extractable: false` – it cannot be exported via `crypto.subtle.exportKey`.
 */
export async function getOrCreateWrappingKey(): Promise<CryptoKey> {
  return withKeyDb(async (db) => {
    const existing = await idbGet(db, KEY_ID);
    if (existing) return existing;

    const key = await crypto.subtle.generateKey(
      KEY_ALGO,
      false /* not extractable */,
      ["encrypt", "decrypt"],
    );
    await idbPut(db, KEY_ID, key);
    return key;
  });
}

// ---------------------------------------------------------------------------
// Wrap / unwrap
// ---------------------------------------------------------------------------

/**
 * Encrypts `token` with AES-GCM and returns a blob string prefixed with WRAPPED_BLOB_MAGIC.
 * The nonce is generated freshly with `crypto.getRandomValues` for every call.
 */
export async function wrapToken(token: string): Promise<string> {
  const key = await getOrCreateWrappingKey();
  // Fresh 12-byte random nonce – NEVER reuse. Counter or time values are forbidden.
  const nonce = crypto.getRandomValues(new Uint8Array(NONCE_LENGTH));
  const plaintext = new TextEncoder().encode(token);
  const ciphertext = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: nonce },
    key,
    plaintext,
  );

  // Pack: nonce (12 bytes) || ciphertext
  const packed = new Uint8Array(NONCE_LENGTH + ciphertext.byteLength);
  packed.set(nonce, 0);
  packed.set(new Uint8Array(ciphertext), NONCE_LENGTH);

  // Use reduce instead of spread to avoid stack overflow for large arrays.
  return (
    WRAPPED_BLOB_MAGIC +
    btoa(packed.reduce((s, b) => s + String.fromCharCode(b), ""))
  );
}

/**
 * Decrypts a blob produced by `wrapToken`.
 * If `blob` does not carry the magic prefix (legacy plaintext), it is returned as-is
 * so callers can handle migration without a special branch.
 *
 * Throws `DOMException` (OperationError) when AES-GCM authentication fails
 * (e.g. the IDB key was rotated or data is corrupted). Callers should catch this
 * and call `clearStoredCredential()` followed by prompting for re-login.
 */
export async function unwrapBlob(blob: string): Promise<string> {
  if (!blob.startsWith(WRAPPED_BLOB_MAGIC)) {
    // Legacy plaintext – pass through for migration path.
    return blob;
  }
  const key = await getOrCreateWrappingKey();
  const packed = Uint8Array.from(
    atob(blob.slice(WRAPPED_BLOB_MAGIC.length)),
    (c) => c.charCodeAt(0),
  );
  const nonce = packed.slice(0, NONCE_LENGTH);
  const ciphertext = packed.slice(NONCE_LENGTH);
  const plaintext = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv: nonce },
    key,
    ciphertext,
  );
  return new TextDecoder().decode(plaintext);
}

/**
 * Removes the IDB key (e.g. after the user explicitly logs out, to clean up the key store).
 * Note: calling this without also clearing the credential store will leave an
 * unrecoverable blob on disk.
 */
export async function deleteWrappingKey(): Promise<void> {
  await withKeyDb(
    (db) =>
      new Promise<void>((resolve, reject) => {
        const tx = db.transaction(STORE_NAME, "readwrite");
        const req = tx.objectStore(STORE_NAME).delete(KEY_ID);
        req.onerror = () => reject(req.error);
        tx.onerror = () =>
          reject(tx.error ?? new Error("IndexedDB transaction failed"));
        tx.oncomplete = () => resolve();
      }),
  );
}
