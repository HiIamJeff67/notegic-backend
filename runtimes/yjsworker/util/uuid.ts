export function convertBytesToUUIDString(bytes: Buffer): string | null {
  if (bytes.length !== 16 || bytes.every(byte => byte === 0)) {
    return null;
  }

  const hex = bytes.toString("hex");

  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

export function convertUUIDStringToBytes(uuidString: string): Buffer {
  if (
    !/^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(
      uuidString
    )
  ) {
    throw new Error("invalid UUID");
  }

  return Buffer.from(uuidString.replaceAll("-", ""), "hex");
}
