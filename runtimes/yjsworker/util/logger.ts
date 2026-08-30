export type LogLevel = "debug" | "info" | "warn" | "error";

const colorByLevel: Record<LogLevel, string> = {
  debug: "\u001b[33m",
  info: "\u001b[36m",
  warn: "\u001b[33m",
  error: "\u001b[31m",
};

const resetColor = "\u001b[0m";

export class Logger {
  log(level: LogLevel, message: string, data?: unknown): void {
    const coloredMessage = `${colorByLevel[level]}${message}${resetColor}`;
    console[level](coloredMessage, ...(data === undefined ? [] : [data]));
  }

  info(message: string, data?: unknown): void {
    this.log("info", message, data);
  }

  debug(message: string, data?: unknown): void {
    this.log("debug", message, data);
  }

  warn(message: string, data?: unknown): void {
    this.log("warn", message, data);
  }

  error(message: string, data?: unknown): void {
    this.log("error", message, data);
  }
}
