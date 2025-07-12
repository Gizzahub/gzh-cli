/**
 * TypeScript type definitions for GZH Manager Node.js bindings
 */

/**
 * Platform types supported by GZH Manager
 */
export type PlatformType = 'github' | 'gitlab' | 'gitea' | 'gogs';

/**
 * Clone strategies available
 */
export type CloneStrategy = 'reset' | 'pull' | 'fetch';

/**
 * Log levels
 */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

/**
 * Plugin status types
 */
export type PluginStatus = 'loaded' | 'error' | 'disabled';

/**
 * Service health status
 */
export type HealthStatus = 'healthy' | 'degraded' | 'unhealthy';

/**
 * Base result interface for all operations
 */
export interface BaseResult {
  /** Operation success status */
  success: boolean;
  /** Error message if operation failed */
  error?: string;
  /** Additional metadata */
  metadata?: Record<string, any>;
}

/**
 * Repository information
 */
export interface RepositoryInfo {
  /** Repository name */
  name: string;
  /** Full repository name (owner/repo) */
  fullName: string;
  /** Repository description */
  description?: string;
  /** Repository URL */
  url: string;
  /** Primary language */
  language?: string;
  /** Star count */
  stars: number;
  /** Fork count */
  forks: number;
  /** Repository size in KB */
  size: number;
  /** Whether repository is private */
  private: boolean;
  /** Whether repository is archived */
  archived: boolean;
  /** Whether repository is a fork */
  fork: boolean;
  /** Last update timestamp */
  updatedAt: string;
  /** Repository topics/tags */
  topics: string[];
}

/**
 * Clone operation statistics
 */
export interface CloneStats {
  /** Total repositories found */
  total: number;
  /** Successfully cloned repositories */
  cloned: number;
  /** Failed clone operations */
  failed: number;
  /** Skipped repositories */
  skipped: number;
  /** Operation start time */
  startTime: string;
  /** Operation end time */
  endTime?: string;
  /** Total duration in seconds */
  duration: number;
  /** Average clone time per repository */
  averageCloneTime: number;
}

/**
 * Clone error details
 */
export interface CloneError {
  /** Repository that failed to clone */
  repository: string;
  /** Error message */
  error: string;
  /** Error code if available */
  code?: string;
  /** Timestamp of error */
  timestamp: string;
}

/**
 * Extended clone result with detailed information
 */
export interface ExtendedCloneResult extends CloneStats {
  /** List of successfully cloned repositories */
  clonedRepositories: RepositoryInfo[];
  /** List of failed repositories with error details */
  errors: CloneError[];
  /** List of skipped repositories with reasons */
  skipped: Array<{
    repository: string;
    reason: string;
  }>;
  /** Platform-specific statistics */
  platformStats: Record<PlatformType, CloneStats>;
}

/**
 * Plugin method information
 */
export interface PluginMethod {
  /** Method name */
  name: string;
  /** Method description */
  description: string;
  /** Parameter schema */
  parameters: Record<string, {
    type: string;
    description: string;
    required: boolean;
    default?: any;
  }>;
  /** Return value schema */
  returns: {
    type: string;
    description: string;
  };
}

/**
 * Extended plugin information
 */
export interface ExtendedPluginInfo extends PluginInfo {
  /** Plugin file path */
  path: string;
  /** Plugin configuration */
  config: Record<string, any>;
  /** Plugin dependencies */
  dependencies: string[];
  /** Plugin permissions */
  permissions: string[];
  /** Detailed method information */
  methodDetails: PluginMethod[];
  /** Plugin load time */
  loadTime: string;
  /** Last execution time */
  lastExecution?: string;
  /** Execution count */
  executionCount: number;
}

/**
 * System metrics information
 */
export interface SystemMetrics {
  /** CPU usage percentage */
  cpuUsage: number;
  /** Memory usage in bytes */
  memoryUsage: number;
  /** Memory usage percentage */
  memoryPercent: number;
  /** Disk usage in bytes */
  diskUsage: number;
  /** Disk usage percentage */
  diskPercent: number;
  /** Network statistics */
  network: {
    bytesReceived: number;
    bytesSent: number;
    packetsReceived: number;
    packetsSent: number;
  };
  /** Go runtime statistics */
  runtime: {
    goroutines: number;
    gcCycles: number;
    heapObjects: number;
    heapSize: number;
  };
}

/**
 * Extended health information
 */
export interface ExtendedHealthInfo extends HealthInfo {
  /** System metrics */
  metrics: SystemMetrics;
  /** Component health status */
  components: Record<string, {
    status: HealthStatus;
    message?: string;
    lastCheck: string;
  }>;
  /** Configuration summary */
  configuration: {
    pluginsEnabled: boolean;
    logLevel: LogLevel;
    timeout: number;
    retryCount: number;
  };
}

/**
 * Bulk clone progress information
 */
export interface CloneProgress {
  /** Current phase */
  phase: 'discovery' | 'filtering' | 'cloning' | 'completed';
  /** Progress percentage */
  percentage: number;
  /** Current repository being processed */
  currentRepository?: string;
  /** Number of repositories processed */
  processed: number;
  /** Total repositories to process */
  total: number;
  /** Estimated time remaining in seconds */
  estimatedTimeRemaining?: number;
  /** Current clone rate (repos/minute) */
  cloneRate: number;
}

/**
 * Configuration validation result
 */
export interface ValidationResult {
  /** Whether configuration is valid */
  valid: boolean;
  /** Validation errors */
  errors: Array<{
    field: string;
    message: string;
    code: string;
  }>;
  /** Validation warnings */
  warnings: Array<{
    field: string;
    message: string;
    code: string;
  }>;
}

/**
 * Platform connectivity test result
 */
export interface ConnectivityTest {
  /** Platform type */
  platform: PlatformType;
  /** Test success status */
  success: boolean;
  /** Response time in milliseconds */
  responseTime: number;
  /** Error message if test failed */
  error?: string;
  /** Additional test details */
  details: {
    url: string;
    statusCode?: number;
    authenticated: boolean;
    apiVersion?: string;
    rateLimitRemaining?: number;
  };
}

/**
 * Event callback types
 */
export type ProgressCallback = (progress: CloneProgress) => void;
export type ErrorCallback = (error: CloneError) => void;
export type CompletionCallback = (result: ExtendedCloneResult) => void;

/**
 * Event emitter interface for clone operations
 */
export interface CloneEventEmitter {
  on(event: 'progress', callback: ProgressCallback): void;
  on(event: 'error', callback: ErrorCallback): void;
  on(event: 'complete', callback: CompletionCallback): void;
  off(event: string, callback: Function): void;
  emit(event: string, ...args: any[]): void;
}

/**
 * Utility type for making properties optional
 */
export type Partial<T> = {
  [P in keyof T]?: T[P];
};

/**
 * Utility type for making properties required
 */
export type Required<T> = {
  [P in keyof T]-?: T[P];
};

/**
 * Utility type for extracting promise type
 */
export type PromiseType<T> = T extends Promise<infer U> ? U : T;

/**
 * Type guard for checking if result is successful
 */
export function isSuccessResult<T>(
  result: OperationResult<T>
): result is OperationResult<T> & { success: true; data: T } {
  return result.success === true && result.data !== undefined;
}

/**
 * Type guard for checking if result is an error
 */
export function isErrorResult<T>(
  result: OperationResult<T>
): result is OperationResult<T> & { success: false; error: string } {
  return result.success === false && result.error !== undefined;
}

// Re-export main types for convenience
export type {
  ClientConfig,
  PlatformConfig,
  CloneFilters,
  BulkCloneRequest,
  PluginExecuteRequest,
  OperationResult,
  CloneResult,
  PluginInfo,
  HealthInfo,
} from './index';