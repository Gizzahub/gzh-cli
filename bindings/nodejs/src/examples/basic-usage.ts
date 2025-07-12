import { Client, createClient, PlatformConfig, BulkCloneRequest } from '../index';

/**
 * Basic usage example for GZH Manager Node.js bindings
 */
async function basicUsageExample() {
  console.log('🚀 GZH Manager Node.js Bindings - Basic Usage Example\n');

  // Create a new client with custom configuration
  const client = createClient({
    timeout: 60,
    retryCount: 3,
    enablePlugins: true,
    logLevel: 'info',
  });

  try {
    // Check client health
    console.log('📊 Checking client health...');
    const healthResult = await client.health();
    if (healthResult.success && healthResult.data) {
      console.log(`✅ Client is ${healthResult.data.status}`);
      console.log(`📝 Version: ${healthResult.data.version}`);
      console.log(`🔌 Plugins: ${healthResult.data.pluginsEnabled ? 'Enabled' : 'Disabled'}\n`);
    } else {
      console.error(`❌ Health check failed: ${healthResult.error}\n`);
    }

    // List available plugins
    console.log('🔌 Listing available plugins...');
    const pluginsResult = await client.listPlugins();
    if (pluginsResult.success && pluginsResult.data) {
      if (pluginsResult.data.length > 0) {
        console.log(`✅ Found ${pluginsResult.data.length} plugins:`);
        pluginsResult.data.forEach(plugin => {
          console.log(`  - ${plugin.name} v${plugin.version} (${plugin.status})`);
        });
      } else {
        console.log('ℹ️  No plugins available');
      }
      console.log();
    } else {
      console.error(`❌ Failed to list plugins: ${pluginsResult.error}\n`);
    }

    // Example bulk clone configuration
    const platforms: PlatformConfig[] = [
      {
        type: 'github',
        token: process.env.GITHUB_TOKEN,
        organizations: ['octocat', 'github'],
        skipArchived: true,
        skipForked: true,
      },
    ];

    const cloneRequest: BulkCloneRequest = {
      platforms,
      outputDir: './cloned-repos',
      concurrency: 3,
      strategy: 'reset',
      includePrivate: false,
      filters: {
        minStars: 10,
        maxSize: 500, // 500MB max
        languages: ['TypeScript', 'JavaScript', 'Go'],
        exclude: ['archived-*', 'deprecated-*'],
      },
    };

    // Perform bulk clone (only if token is available)
    if (process.env.GITHUB_TOKEN) {
      console.log('📥 Starting bulk clone operation...');
      const cloneResult = await client.bulkClone(cloneRequest);
      
      if (cloneResult.success && cloneResult.data) {
        const data = cloneResult.data;
        console.log(`✅ Clone operation completed:`);
        console.log(`  📁 Total repositories: ${data.total}`);
        console.log(`  ✅ Successfully cloned: ${data.cloned}`);
        console.log(`  ❌ Failed: ${data.failed}`);
        console.log(`  ⏭️  Skipped: ${data.skipped}`);
        console.log(`  ⏱️  Duration: ${data.duration}s`);
        
        if (data.errors.length > 0) {
          console.log(`\n❌ Errors encountered:`);
          data.errors.forEach(error => {
            console.log(`  - ${error.repository}: ${error.error}`);
          });
        }
      } else {
        console.error(`❌ Clone operation failed: ${cloneResult.error}`);
      }
    } else {
      console.log('ℹ️  Skipping bulk clone (GITHUB_TOKEN not set)');
    }

    console.log('\n🎉 Example completed successfully!');

  } catch (error) {
    console.error('💥 An error occurred:', error);
  } finally {
    // Always clean up resources
    client.destroy();
    console.log('🧹 Client resources cleaned up');
  }
}

/**
 * Plugin execution example
 */
async function pluginExecutionExample() {
  console.log('\n🔌 Plugin Execution Example\n');

  const client = createClient();

  try {
    // Execute a hypothetical plugin
    const pluginResult = await client.executePlugin({
      pluginName: 'repository-analyzer',
      method: 'analyze',
      args: {
        path: './some-repo',
        includeMetrics: true,
      },
      timeout: 30,
    });

    if (pluginResult.success) {
      console.log('✅ Plugin executed successfully:', pluginResult.data);
    } else {
      console.error('❌ Plugin execution failed:', pluginResult.error);
    }

  } catch (error) {
    console.error('💥 Plugin execution error:', error);
  } finally {
    client.destroy();
  }
}

/**
 * Error handling example
 */
async function errorHandlingExample() {
  console.log('\n🛡️  Error Handling Example\n');

  const client = createClient({
    timeout: 1, // Very short timeout to trigger errors
  });

  try {
    // This will likely timeout
    const result = await client.bulkClone({
      platforms: [{
        type: 'github',
        organizations: ['nonexistent-org'],
      }],
      outputDir: './test-output',
    });

    if (!result.success) {
      console.log('✅ Error handled gracefully:', result.error);
    }

  } catch (error) {
    console.log('✅ Exception caught and handled:', error.message);
  } finally {
    client.destroy();
  }
}

// Run examples if this file is executed directly
if (require.main === module) {
  (async () => {
    await basicUsageExample();
    await pluginExecutionExample();
    await errorHandlingExample();
  })().catch(console.error);
}

export {
  basicUsageExample,
  pluginExecutionExample,
  errorHandlingExample,
};