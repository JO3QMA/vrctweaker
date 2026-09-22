export type AffinityParseResult =
  { ok: true; mask: bigint } | { ok: false; raw: string };

const HEX_RE = /^[0-9a-fA-F]*$/;

export function parseAffinityHex(raw: string): AffinityParseResult {
  let s = raw.trim();
  if (s.startsWith("0x") || s.startsWith("0X")) {
    s = s.slice(2);
  }
  if (s === "") {
    return { ok: true, mask: 0n };
  }
  if (!HEX_RE.test(s)) {
    return { ok: false, raw };
  }
  return { ok: true, mask: BigInt(`0x${s}`) };
}

/** Shortest uppercase hex for VRChat; zero mask → empty string. */
export function formatAffinityHex(mask: bigint): string {
  if (mask === 0n) {
    return "";
  }
  return mask.toString(16).toUpperCase();
}

export function isValidAffinityMask(mask: bigint): boolean {
  return mask > 0n;
}

export function allCoresAllowedMask(coreCount: number): bigint {
  if (coreCount <= 0) {
    return 0n;
  }
  return (1n << BigInt(coreCount)) - 1n;
}

export function coreStatesFromMask(mask: bigint, coreCount: number): boolean[] {
  const out: boolean[] = [];
  for (let i = 0; i < coreCount; i++) {
    out.push((mask & (1n << BigInt(i))) !== 0n);
  }
  return out;
}

export function maskFromCoreStates(
  states: boolean[],
  hiddenMask: bigint,
): bigint {
  let visible = 0n;
  for (let i = 0; i < states.length; i++) {
    if (states[i]) {
      visible |= 1n << BigInt(i);
    }
  }
  return hiddenMask | visible;
}

export function hiddenMaskAbove(mask: bigint, coreCount: number): bigint {
  if (coreCount <= 0) {
    return mask;
  }
  const keep = allCoresAllowedMask(coreCount);
  return mask & ~keep;
}

export function hasHiddenBitsAbove(mask: bigint, coreCount: number): boolean {
  return hiddenMaskAbove(mask, coreCount) !== 0n;
}

export function mergeVisibleWithHidden(
  states: boolean[],
  hiddenMask: bigint,
): bigint {
  return maskFromCoreStates(states, hiddenMask);
}
