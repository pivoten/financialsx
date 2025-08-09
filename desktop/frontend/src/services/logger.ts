/**
 * Frontend Logging Service
 * Replaces console.log with structured logging that can be sent to backend
 */

import { 
  InitializeLogging, 
  LogMessage, 
  SetDebugMode, 
  GetDebugMode,
  GetLogFilePath 
} from '../../wailsjs/go/main/App'

export type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'FATAL'

interface LogData {
  [key: string]: any
}

class Logger {
  private static instance: Logger
  private debugMode: boolean = false
  private initialized: boolean = false
  private component: string = 'Frontend'
  private queue: Array<{level: LogLevel, message: string, data?: LogData}> = []

  private constructor() {
    // Initialize on first use
    this.initialize()
  }

  static getInstance(): Logger {
    if (!Logger.instance) {
      Logger.instance = new Logger()
    }
    return Logger.instance
  }

  async initialize(): Promise<void> {
    if (this.initialized) return

    try {
      // Check if debug mode is stored in localStorage
      const storedDebugMode = localStorage.getItem('debugMode') === 'true'
      
      // Initialize backend logging
      const result = await InitializeLogging(storedDebugMode)
      
      if (result.success) {
        this.debugMode = result.debugMode
        this.initialized = true
        
        // Process any queued logs
        this.processQueue()
      }
    } catch (error) {
      // Fallback to console if backend is not available
      console.error('Failed to initialize logging:', error)
    }
  }

  private async processQueue(): Promise<void> {
    while (this.queue.length > 0) {
      const log = this.queue.shift()
      if (log) {
        await this.log(log.level, log.message, log.data)
      }
    }
  }

  setComponent(component: string): void {
    this.component = component
  }

  private async log(level: LogLevel, message: string, data?: LogData): Promise<void> {
    // If not initialized, queue the log
    if (!this.initialized) {
      this.queue.push({level, message, data})
      return
    }

    // In development, also log to console
    if (import.meta.env.DEV || this.debugMode) {
      const consoleMethod = this.getConsoleMethod(level)
      if (data) {
        console[consoleMethod](`[${level}] ${message}`, data)
      } else {
        console[consoleMethod](`[${level}] ${message}`)
      }
    }

    // Skip DEBUG logs if not in debug mode
    if (level === 'DEBUG' && !this.debugMode) {
      return
    }

    try {
      // Send to backend
      await LogMessage(level, message, this.component, data || {})
    } catch (error) {
      // Fallback to console on error
      console.error('Failed to send log to backend:', error)
    }
  }

  private getConsoleMethod(level: LogLevel): 'log' | 'info' | 'warn' | 'error' {
    switch (level) {
      case 'DEBUG':
      case 'INFO':
        return 'log'
      case 'WARN':
        return 'warn'
      case 'ERROR':
      case 'FATAL':
        return 'error'
      default:
        return 'log'
    }
  }

  // Public logging methods
  debug(message: string, data?: LogData): void {
    this.log('DEBUG', message, data)
  }

  info(message: string, data?: LogData): void {
    this.log('INFO', message, data)
  }

  warn(message: string, data?: LogData): void {
    this.log('WARN', message, data)
  }

  error(message: string, data?: LogData): void {
    this.log('ERROR', message, data)
  }

  fatal(message: string, data?: LogData): void {
    this.log('FATAL', message, data)
  }

  // Log with timing
  time(label: string): void {
    const startTime = performance.now()
    this.debug(`Timer started: ${label}`, { startTime })
  }

  timeEnd(label: string): void {
    const endTime = performance.now()
    this.debug(`Timer ended: ${label}`, { endTime })
  }

  // Log method entry/exit for debugging
  methodEntry(methodName: string, args?: any): void {
    this.debug(`Entering ${methodName}`, { args })
  }

  methodExit(methodName: string, result?: any): void {
    this.debug(`Exiting ${methodName}`, { result })
  }

  // Log API calls
  apiCall(method: string, endpoint: string, data?: any): void {
    this.debug(`API Call: ${method} ${endpoint}`, { data })
  }

  apiResponse(method: string, endpoint: string, response?: any, error?: any): void {
    if (error) {
      this.error(`API Error: ${method} ${endpoint}`, { error })
    } else {
      this.debug(`API Response: ${method} ${endpoint}`, { response })
    }
  }

  // Set debug mode
  async setDebugMode(enabled: boolean): Promise<void> {
    try {
      const result = await SetDebugMode(enabled)
      if (result.success) {
        this.debugMode = enabled
        localStorage.setItem('debugMode', enabled.toString())
        this.info(`Debug mode ${enabled ? 'enabled' : 'disabled'}`)
      }
    } catch (error) {
      this.error('Failed to set debug mode', { error })
    }
  }

  // Get debug mode status
  async getDebugMode(): Promise<boolean> {
    try {
      const debugMode = await GetDebugMode()
      this.debugMode = debugMode
      return debugMode
    } catch (error) {
      this.error('Failed to get debug mode', { error })
      return false
    }
  }

  // Get log file path
  async getLogFilePath(): Promise<string> {
    try {
      return await GetLogFilePath()
    } catch (error) {
      this.error('Failed to get log file path', { error })
      return ''
    }
  }
}

// Export singleton instance
const logger = Logger.getInstance()
export default logger

// Export convenience functions that match console API
export const log = {
  debug: (message: string, data?: LogData) => logger.debug(message, data),
  info: (message: string, data?: LogData) => logger.info(message, data),
  warn: (message: string, data?: LogData) => logger.warn(message, data),
  error: (message: string, data?: LogData) => logger.error(message, data),
  fatal: (message: string, data?: LogData) => logger.fatal(message, data),
}