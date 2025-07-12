import { Client, PlatformConfig, BulkCloneRequest, CloneFilters } from '../index';
import { ExtendedCloneResult, ConnectivityTest, ValidationResult } from '../types';

/**
 * Advanced usage example with multiple platforms and complex filtering
 */
async function multiPlatformCloneExample() {
  console.log('🌐 Multi-Platform Clone Example\n');

  const client = new Client({
    timeout: 120,
    retryCount: 5,
    enablePlugins: true,
    logLevel: 'debug',
  });

  try {
    // Configure multiple platforms
    const platforms: PlatformConfig[] = [
      // GitHub configuration
      {
        type: 'github',
        token: process.env.GITHUB_TOKEN,
        organizations: ['microsoft', 'google', 'facebook'],
        skipArchived: true,
        skipForked: true,
      },
      // GitLab configuration
      {
        type: 'gitlab',
        baseUrl: 'https://gitlab.com',
        token: process.env.GITLAB_TOKEN,
        organizations: ['gitlab-org', 'gitlab-examples'],
        skipArchived: true,
      },
      // Self-hosted Gitea instance
      {
        type: 'gitea',
        baseUrl: 'https://gitea.example.com',
        token: process.env.GITEA_TOKEN,
        organizations: ['my-org'],
      },
    ];

    // Advanced filtering configuration
    const filters: CloneFilters = {
      include: ['*-api', '*-sdk', '*-library'],
      exclude: ['*-test*', '*-demo*', '*-example*'],
      minStars: 100,
      maxSize: 1000, // 1GB max
      languages: ['Go', 'TypeScript', 'Python', 'Rust'],
      updatedAfter: '2023-01-01',
    };

    const cloneRequest: BulkCloneRequest = {
      platforms,
      outputDir: './multi-platform-repos',
      concurrency: 10,
      strategy: 'pull',
      includePrivate: false,
      filters,
    };

    console.log('🔍 Validating configuration...');
    const validation = await validateConfiguration(cloneRequest);
    if (!validation.valid) {
      console.error('❌ Configuration validation failed:');
      validation.errors.forEach(error => {
        console.error(`  - ${error.field}: ${error.message}`);
      });
      return;
    }

    console.log('🌐 Testing platform connectivity...');
    const connectivityTests = await testPlatformConnectivity(platforms);
    const failedTests = connectivityTests.filter(test => !test.success);
    
    if (failedTests.length > 0) {
      console.warn('⚠️  Some platforms are not accessible:');
      failedTests.forEach(test => {
        console.warn(`  - ${test.platform}: ${test.error}`);
      });
    }

    console.log('📥 Starting multi-platform clone...');
    const result = await client.bulkClone(cloneRequest);

    if (result.success && result.data) {
      displayCloneResults(result.data);
    } else {
      console.error(`❌ Clone failed: ${result.error}`);
    }

  } catch (error) {
    console.error('💥 Error in multi-platform clone:', error);
  } finally {
    client.destroy();
  }
}

/**
 * Example with custom progress tracking
 */
async function progressTrackingExample() {
  console.log('\n📊 Progress Tracking Example\n');

  const client = new Client();

  try {
    // Simulate a clone operation with progress tracking
    console.log('🚀 Starting clone with progress tracking...');
    
    const platforms: PlatformConfig[] = [{
      type: 'github',
      token: process.env.GITHUB_TOKEN,
      organizations: ['nodejs'],
    }];

    // Start the clone operation
    const clonePromise = client.bulkClone({
      platforms,
      outputDir: './progress-tracked-repos',
      concurrency: 5,
    });

    // Simulate progress updates (in real implementation, this would come from the Go library)
    const progressInterval = setInterval(() => {
      const progress = Math.floor(Math.random() * 100);
      console.log(`📊 Progress: ${progress}%`);
    }, 2000);

    const result = await clonePromise;
    clearInterval(progressInterval);

    if (result.success) {
      console.log('✅ Clone completed with progress tracking');
    }

  } catch (error) {
    console.error('💥 Progress tracking error:', error);
  } finally {
    client.destroy();
  }
}

/**
 * Plugin development and testing example
 */
async function pluginDevelopmentExample() {
  console.log('\n🔧 Plugin Development Example\n');

  const client = new Client({
    enablePlugins: true,
    pluginDir: './custom-plugins',
  });

  try {
    // List available plugins
    const pluginsResult = await client.listPlugins();
    if (pluginsResult.success && pluginsResult.data) {
      console.log('🔌 Available plugins:');
      pluginsResult.data.forEach(plugin => {
        console.log(`  - ${plugin.name}: ${plugin.description}`);
        console.log(`    Methods: ${plugin.methods.join(', ')}`);
      });
    }

    // Execute a custom analysis plugin
    const analysisResult = await client.executePlugin({
      pluginName: 'code-quality-analyzer',
      method: 'analyze_repository',
      args: {
        repository_path: './some-repo',
        include_metrics: ['complexity', 'coverage', 'duplication'],
        output_format: 'json',
      },
      timeout: 60,
    });

    if (analysisResult.success) {
      console.log('📊 Code quality analysis:', analysisResult.data);
    }

    // Execute a security scanning plugin
    const securityResult = await client.executePlugin({
      pluginName: 'security-scanner',
      method: 'scan_dependencies',
      args: {
        scan_depth: 'deep',
        include_dev_dependencies: false,
        vulnerability_threshold: 'medium',
      },
    });

    if (securityResult.success) {
      console.log('🔒 Security scan results:', securityResult.data);
    }

  } catch (error) {
    console.error('💥 Plugin development error:', error);
  } finally {
    client.destroy();
  }
}

/**
 * Batch operations example
 */
async function batchOperationsExample() {
  console.log('\n⚡ Batch Operations Example\n');

  const client = new Client({
    timeout: 300, // 5 minutes for large operations
  });

  try {
    // Perform multiple operations in sequence
    const operations = [
      {
        name: 'Open Source Projects',
        platforms: [{
          type: 'github' as const,
          organizations: ['apache', 'kubernetes'],
        }],
        outputDir: './open-source',
      },
      {
        name: 'Company Projects',
        platforms: [{
          type: 'gitlab' as const,
          baseUrl: 'https://gitlab.company.com',
          organizations: ['backend-team', 'frontend-team'],
        }],
        outputDir: './company-projects',
      },
    ];

    for (const operation of operations) {
      console.log(`🔄 Processing: ${operation.name}`);
      
      const result = await client.bulkClone({
        platforms: operation.platforms,
        outputDir: operation.outputDir,
        concurrency: 8,
        strategy: 'reset',
      });

      if (result.success && result.data) {
        console.log(`✅ ${operation.name}: ${result.data.cloned} repositories cloned`);
      } else {
        console.error(`❌ ${operation.name} failed: ${result.error}`);
      }
    }

  } catch (error) {
    console.error('💥 Batch operations error:', error);
  } finally {
    client.destroy();
  }
}

// Helper functions

/**
 * Validate clone configuration
 */
async function validateConfiguration(request: BulkCloneRequest): Promise<ValidationResult> {
  const errors: ValidationResult['errors'] = [];
  const warnings: ValidationResult['warnings'] = [];

  // Basic validation
  if (!request.platforms || request.platforms.length === 0) {
    errors.push({
      field: 'platforms',
      message: 'At least one platform must be configured',
      code: 'PLATFORMS_REQUIRED',
    });
  }

  if (!request.outputDir) {
    errors.push({
      field: 'outputDir',
      message: 'Output directory is required',
      code: 'OUTPUT_DIR_REQUIRED',
    });
  }

  // Check for tokens
  for (const platform of request.platforms) {
    if (!platform.token && platform.type !== 'github') {
      warnings.push({
        field: `platforms.${platform.type}.token`,
        message: `No token provided for ${platform.type}, only public repositories will be accessible`,
        code: 'TOKEN_MISSING',
      });
    }
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  };
}

/**
 * Test connectivity to platforms
 */
async function testPlatformConnectivity(platforms: PlatformConfig[]): Promise<ConnectivityTest[]> {
  // In a real implementation, this would make actual HTTP requests
  return platforms.map(platform => ({
    platform: platform.type,
    success: Math.random() > 0.2, // 80% success rate for demo
    responseTime: Math.floor(Math.random() * 1000) + 100,
    error: Math.random() > 0.8 ? 'Connection timeout' : undefined,
    details: {
      url: platform.baseUrl || `https://${platform.type}.com`,
      statusCode: 200,
      authenticated: !!platform.token,
      apiVersion: '4.0',
      rateLimitRemaining: Math.floor(Math.random() * 5000),
    },
  }));
}

/**
 * Display detailed clone results
 */
function displayCloneResults(result: any) {
  console.log('\n📊 Clone Results Summary:');
  console.log(`  📁 Total repositories: ${result.total}`);
  console.log(`  ✅ Successfully cloned: ${result.cloned}`);
  console.log(`  ❌ Failed: ${result.failed}`);
  console.log(`  ⏭️  Skipped: ${result.skipped}`);
  console.log(`  ⏱️  Duration: ${result.duration}s`);

  if (result.errors && result.errors.length > 0) {
    console.log('\n❌ Failed repositories:');
    result.errors.slice(0, 5).forEach((error: any) => {
      console.log(`  - ${error.repository}: ${error.error}`);
    });
    
    if (result.errors.length > 5) {
      console.log(`  ... and ${result.errors.length - 5} more`);
    }
  }
}

// Export examples
export {
  multiPlatformCloneExample,
  progressTrackingExample,
  pluginDevelopmentExample,
  batchOperationsExample,
};

// Run examples if this file is executed directly
if (require.main === module) {
  (async () => {
    await multiPlatformCloneExample();
    await progressTrackingExample();
    await pluginDevelopmentExample();
    await batchOperationsExample();
  })().catch(console.error);
}