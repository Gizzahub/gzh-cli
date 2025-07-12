import * as path from 'path';
import * as bindings from 'bindings';

// Load the native addon
const addon = bindings('gzh_manager_native');

/**
 * Client configuration options
 */
export interface ClientConfig {
  /** Request timeout in seconds */
  timeout?: number;
  /** Number of retry attempts */
  retryCount?: number;
  /** Enable plugin system */
  enablePlugins?: boolean;
  /** Plugin directory path */
  pluginDir?: string;
  /** Log level (debug, info, warn, error) */
  logLevel?: string;
  /** Log file path */
  logFile?: string;
}

/**
 * Platform configuration for bulk clone operations
 */
export interface PlatformConfig {
  /** Platform type (github, gitlab, gitea, gogs) */
  type: 'github' | 'gitlab' | 'gitea' | 'gogs';
  /** Base URL for the platform */
  baseUrl?: string;
  /** Authentication token */
  token?: string;
  /** Organizations/groups to clone */
  organizations: string[];
  /** Skip archived repositories */
  skipArchived?: boolean;
  /** Skip forked repositories */
  skipForked?: boolean;
}

/**
 * Clone filters for filtering repositories
 */
export interface CloneFilters {
  /** Include only repositories matching these patterns */
  include?: string[];
  /** Exclude repositories matching these patterns */
  exclude?: string[];
  /** Minimum star count */
  minStars?: number;
  /** Maximum repository size in MB */
  maxSize?: number;
  /** Include only repositories with these languages */
  languages?: string[];
  /** Include only repositories updated after this date */
  updatedAfter?: string;
}

/**
 * Bulk clone request configuration
 */
export interface BulkCloneRequest {
  /** Platform configurations */
  platforms: PlatformConfig[];
  /** Output directory for cloned repositories */
  outputDir: string;
  /** Concurrency level for clone operations */
  concurrency?: number;
  /** Clone strategy (reset, pull, fetch) */
  strategy?: 'reset' | 'pull' | 'fetch';
  /** Include private repositories */
  includePrivate?: boolean;
  /** Repository filters */
  filters?: CloneFilters;
}

/**
 * Plugin execution request
 */
export interface PluginExecuteRequest {
  /** Plugin name */
  pluginName: string;
  /** Method to execute */
  method: string;
  /** Method arguments */
  args?: Record<string, any>;
  /** Execution timeout in seconds */
  timeout?: number;
}

/**
 * Operation result
 */
export interface OperationResult<T = any> {
  /** Operation success status */
  success: boolean;
  /** Error message if operation failed */
  error?: string;
  /** Result data if operation succeeded */
  data?: T;
}

/**
 * Clone operation result
 */
export interface CloneResult {
  /** Number of repositories successfully cloned */
  cloned: number;
  /** Number of repositories that failed to clone */
  failed: number;
  /** Number of repositories skipped */
  skipped: number;
  /** Total number of repositories processed */
  total: number;
  /** List of failed repositories with error messages */
  errors: Array<{
    repository: string;
    error: string;
  }>;
  /** Execution time in seconds */
  duration: number;
}

/**
 * Plugin information
 */
export interface PluginInfo {
  /** Plugin name */
  name: string;
  /** Plugin version */
  version: string;
  /** Plugin description */
  description: string;
  /** Plugin author */
  author: string;
  /** Available methods */
  methods: string[];
  /** Plugin status */
  status: 'loaded' | 'error' | 'disabled';
}

/**
 * Health check result
 */
export interface HealthInfo {
  /** Service status */
  status: 'healthy' | 'degraded' | 'unhealthy';
  /** Version information */
  version: string;
  /** System uptime in seconds */
  uptime: number;
  /** Plugin system status */
  pluginsEnabled: boolean;
  /** Number of loaded plugins */
  pluginCount: number;
  /** Last check timestamp */
  timestamp: string;
}

/**
 * Default client configuration
 */
const DEFAULT_CONFIG: ClientConfig = {
  timeout: 30,
  retryCount: 3,
  enablePlugins: true,
  logLevel: 'info',
};

/**
 * GZH Manager client for Node.js
 */
export class Client {
  private clientId: number;
  private config: ClientConfig;

  /**
   * Create a new GZH Manager client
   * @param config Client configuration options
   */
  constructor(config: ClientConfig = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.clientId = addon.createClient(this.config);
    
    if (this.clientId < 0) {
      throw new Error('Failed to create GZH Manager client');
    }
  }

  /**
   * Destroy the client and free resources
   */
  destroy(): void {
    if (this.clientId >= 0) {
      addon.destroyClient(this.clientId);
      this.clientId = -1;
    }
  }

  /**
   * Perform bulk clone operation
   * @param request Bulk clone request configuration
   * @returns Promise resolving to clone operation result
   */
  async bulkClone(request: BulkCloneRequest): Promise<OperationResult<CloneResult>> {
    this.validateClientId();
    
    const requestData = {
      platforms: JSON.stringify(request.platforms),
      outputDir: request.outputDir,
      concurrency: request.concurrency || 5,
      strategy: request.strategy || 'reset',
      includePrivate: request.includePrivate || false,
      filters: request.filters ? JSON.stringify(request.filters) : undefined,
    };

    return new Promise((resolve, reject) => {
      try {
        const result = addon.bulkClone(this.clientId, requestData);
        const parsedResult = this.parseResult<CloneResult>(result);
        resolve(parsedResult);
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * List available plugins
   * @returns Promise resolving to list of plugin information
   */
  async listPlugins(): Promise<OperationResult<PluginInfo[]>> {
    this.validateClientId();

    return new Promise((resolve, reject) => {
      try {
        const result = addon.listPlugins(this.clientId);
        const parsedResult = this.parseResult<PluginInfo[]>(result);
        resolve(parsedResult);
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Execute a plugin method
   * @param request Plugin execution request
   * @returns Promise resolving to plugin execution result
   */
  async executePlugin(request: PluginExecuteRequest): Promise<OperationResult<any>> {
    this.validateClientId();

    const argsJson = request.args ? JSON.stringify(request.args) : '{}';
    const timeout = request.timeout || 30;

    return new Promise((resolve, reject) => {
      try {
        const result = addon.executePlugin(
          this.clientId,
          request.pluginName,
          request.method,
          argsJson,
          timeout
        );
        const parsedResult = this.parseResult(result);
        resolve(parsedResult);
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Get client health information
   * @returns Promise resolving to health information
   */
  async health(): Promise<OperationResult<HealthInfo>> {
    this.validateClientId();

    return new Promise((resolve, reject) => {
      try {
        const result = addon.health(this.clientId);
        const parsedResult = this.parseResult<HealthInfo>(result);
        resolve(parsedResult);
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Get client configuration
   * @returns Current client configuration
   */
  getConfig(): Readonly<ClientConfig> {
    return { ...this.config };
  }

  /**
   * Check if client is valid
   * @returns True if client is valid
   */
  isValid(): boolean {
    return this.clientId >= 0;
  }

  private validateClientId(): void {
    if (this.clientId < 0) {
      throw new Error('Client has been destroyed or is invalid');
    }
  }

  private parseResult<T>(result: any): OperationResult<T> {
    if (!result.success) {
      return {
        success: false,
        error: result.error || 'Unknown error occurred',
      };
    }

    let data: T | undefined;
    if (result.data) {
      try {
        data = JSON.parse(result.data);
      } catch (error) {
        return {
          success: false,
          error: `Failed to parse result data: ${error.message}`,
        };
      }
    }

    return {
      success: true,
      data,
    };
  }
}

/**
 * Create a new GZH Manager client with the given configuration
 * @param config Client configuration options
 * @returns New client instance
 */
export function createClient(config?: ClientConfig): Client {
  return new Client(config);
}

/**
 * Get default client configuration
 * @returns Default configuration object
 */
export function getDefaultConfig(): ClientConfig {
  return { ...DEFAULT_CONFIG };
}

// Export types and constants
export {
  DEFAULT_CONFIG as defaultConfig,
};

// Default export
export default {
  Client,
  createClient,
  getDefaultConfig,
};