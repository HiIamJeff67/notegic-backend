export function required(name: string): string {
  const value = process.env[name]?.trim();
  if (value === undefined || value === "") {
    throw new Error(`${name} is required`);
  }

  return value;
}

export function positiveInteger(name: string, fallback: number): number {
  const value = Number(process.env[name] ?? fallback);
  if (!Number.isInteger(value) || value <= 0) {
    throw new Error(`${name} must be a positive integer`);
  }

  return value;
}

export function port(name: string, fallback?: number): number {
  const rawValue = process.env[name];
  const value =
    fallback !== undefined && (rawValue === undefined || rawValue === "")
      ? fallback
      : fallback === undefined
        ? Number(required(name))
        : Number(rawValue);
  if (!Number.isInteger(value) || value < 1 || value > 65535) {
    throw new Error(`${name} must be an integer between 1 and 65535`);
  }

  return value;
}
