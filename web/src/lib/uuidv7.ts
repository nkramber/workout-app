// uuidv7 makes a UUID of version 7 (RFC 9562, section 5.7): 48 bits of
// Unix time in milliseconds, then 74 random bits. Each change gets one as
// its client op id, so the server of Phase 6 can apply a repeated op once
// (REC-1, D-132). The time comes first, so the ids of one phone sort in
// the order of their minutes and seconds.
export function uuidv7(now: number = Date.now(), random: (b: Uint8Array) => void = (b) => crypto.getRandomValues(b)): string {
  const bytes = new Uint8Array(16);
  random(bytes);
  let ms = Math.floor(now);
  for (let i = 5; i >= 0; i--) {
    bytes[i] = ms % 256;
    ms = Math.floor(ms / 256);
  }
  bytes[6] = 0x70 | (bytes[6] & 0x0f); // version 7
  bytes[8] = 0x80 | (bytes[8] & 0x3f); // variant 10
  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}
