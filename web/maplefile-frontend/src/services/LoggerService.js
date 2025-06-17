// This service handles logging throughout your app
// It's simple but will help us understand how services work
export class LoggerService {
  constructor() {
    this.logs = [];
  }

  // Log a message with a timestamp
  log(message) {
    const timestamp = new Date().toLocaleTimeString();
    const logEntry = `[${timestamp}] ${message}`;
    this.logs.push(logEntry);
    console.log(logEntry); // Also show in browser console
  }

  // Get all logged messages
  getLogs() {
    return this.logs;
  }

  // Clear all logs
  clearLogs() {
    this.logs = [];
    this.log("LoggerService: Logs cleared");
  }
}
