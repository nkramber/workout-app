// uuidv7 makes a UUID of version 7 (RFC 9562, section 5.7): 48 bits of
// Unix time in milliseconds, then 74 random bits. Each change gets one as
// its client op id, so the server of Phase 6 can apply a repeated op once
// (REC-1, D-132). The time comes first, so the ids of one phone sort in
// the order of their minutes and seconds.
export function uuidv7(now: number = Date.now(), random: (b: Uint8Array) => void = (b) => crypto.getRandomValues(b)): string {
  return build(Math.floor(now), null, random);
}

// monotonicUuidv7 gives a generator of UUIDv7 ids that rise strictly, in
// the order of the calls. Two ids of the same millisecond get a counter
// in the 12 bits of rand_a, the method 1 of RFC 9562, section 6.2. A
// clock that goes back keeps the last time. The outbox sorts its entries
// by their time, then by their op id, so two changes of one millisecond
// keep their order (D-132).
export function monotonicUuidv7(
  random: (b: Uint8Array) => void = (b) => crypto.getRandomValues(b),
): (now?: number) => string {
  let lastMs = -1;
  let counter = 0;
  return (now = Date.now()) => {
    let ms = Math.floor(now);
    if (ms <= lastMs) {
      ms = lastMs;
      counter++;
      if (counter > 0xfff) {
        ms++;
        counter = 0;
      }
    } else {
      counter = 0;
    }
    lastMs = ms;
    return build(ms, counter, random);
  };
}

// nextId is the one generator of the ids of the phone: each op id and
// each entity id.
export const nextId = monotonicUuidv7();

function build(ms: number, counter: number | null, random: (b: Uint8Array) => void): string {
  const bytes = new Uint8Array(16);
  random(bytes);
  for (let i = 5; i >= 0; i--) {
    bytes[i] = ms % 256;
    ms = Math.floor(ms / 256);
  }
  if (counter !== null) {
    bytes[6] = counter >> 8;
    bytes[7] = counter & 0xff;
  }
  bytes[6] = 0x70 | (bytes[6] & 0x0f); // version 7
  bytes[8] = 0x80 | (bytes[8] & 0x3f); // variant 10
  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}
