#!/usr/bin/env python3
"""
GZH Manager Python Client - Basic Usage Example

This example demonstrates basic usage of the GZH Manager Python client.
"""

import os
import sys
from pathlib import Path

# Add the parent directory to Python path so we can import gzhclient
sys.path.insert(0, str(Path(__file__).parent.parent))

import gzhclient
from gzhclient import PlatformConfig, CloneFilters
from datetime import datetime, timedelta


def main():
    """Demonstrate basic GZH Manager client usage."""
    
    print("🚀 GZH Manager Python Client - Basic Usage Example")
    print("=" * 50)
    
    # Create client with default configuration
    print("\n1. Creating client...")
    with gzhclient.Client() as client:
        print("✅ Client created successfully")
        
        # Check client health
        print("\n2. Checking client health...")
        health = client.health()
        print(f"📊 Overall status: {health.overall.value}")
        print(f"📊 Components: {len(health.components)}")
        
        for name, component in health.components.items():
            print(f"   - {name}: {component.status.value} - {component.message}")
        
        # Get system metrics
        print("\n3. Getting system metrics...")
        try:
            metrics = client.get_system_metrics()
            print(f"💻 CPU cores: {metrics.cpu.cores}")
            print(f"💻 CPU usage: {metrics.cpu.usage:.1f}%")
            print(f"💾 Memory total: {metrics.memory.total // (1024**3)} GB")
            print(f"💾 Memory usage: {metrics.memory.usage:.1f}%")
            print(f"💽 Disk usage: {metrics.disk.usage:.1f}%")
            print(f"⏱️  System uptime: {metrics.uptime / 3600:.1f} hours")
        except Exception as e:
            print(f"⚠️  Could not get system metrics: {e}")
        
        # List plugins
        print("\n4. Listing plugins...")
        try:
            plugins = client.list_plugins()
            if plugins:
                print(f"🔌 Found {len(plugins)} plugin(s):")
                for plugin in plugins:
                    print(f"   - {plugin.name} v{plugin.version}: {plugin.description}")
                    print(f"     Status: {plugin.status}, Calls: {plugin.call_count}")
            else:
                print("🔌 No plugins loaded")
        except Exception as e:
            print(f"⚠️  Could not list plugins: {e}")
        
        # Example bulk clone (commented out to avoid actual cloning)
        print("\n5. Bulk clone example (dry run)...")
        print("📂 This example shows how to configure bulk cloning:")
        
        # Configure platforms
        platforms = [
            PlatformConfig(
                type="github",
                token=os.getenv("GITHUB_TOKEN", "your-github-token"),
                organizations=["octocat"],  # Example organization
            ),
            PlatformConfig(
                type="gitlab",
                url="https://gitlab.com",
                token=os.getenv("GITLAB_TOKEN", "your-gitlab-token"),
                organizations=["gitlab-org"],  # Example organization
            )
        ]
        
        # Configure filters
        filters = CloneFilters(
            languages=["python", "go", "javascript"],
            updated_after=datetime.now() - timedelta(days=30),  # Last 30 days
            exclude_repos=["test-repo", "archive-*"]
        )
        
        print(f"   📋 Platforms: {len(platforms)}")
        print(f"   📋 Output directory: ./repositories")
        print(f"   📋 Concurrency: 5")
        print(f"   📋 Strategy: reset")
        print(f"   📋 Include private: False")
        print(f"   📋 Language filters: {', '.join(filters.languages)}")
        print(f"   📋 Updated after: {filters.updated_after.strftime('%Y-%m-%d')}")
        
        # Uncomment to actually perform bulk clone
        # WARNING: This will clone repositories to your local machine
        """
        result = client.bulk_clone(
            platforms=platforms,
            output_dir="./repositories",
            concurrency=5,
            strategy="reset",
            include_private=False,
            filters=filters
        )
        
        print(f"📊 Bulk clone completed:")
        print(f"   Total repositories: {result.total_repos}")
        print(f"   Successfully cloned: {result.success_count}")
        print(f"   Failed: {result.failure_count}")
        print(f"   Skipped: {result.skipped_count}")
        print(f"   Duration: {result.duration:.2f} seconds")
        
        if result.failure_count > 0:
            print(f"❌ Failed repositories:")
            for repo_result in result.results:
                if repo_result.status == "failed":
                    print(f"   - {repo_result.repo_name}: {repo_result.error}")
        """
        
        print("\n6. Example plugin execution (if plugins are available)...")
        try:
            plugins = client.list_plugins()
            if plugins:
                # Execute first plugin as example
                plugin = plugins[0]
                print(f"🔌 Executing plugin: {plugin.name}")
                
                result = client.execute_plugin(
                    plugin_name=plugin.name,
                    method="info",  # Common method name
                    args={"verbose": True},
                    timeout=10
                )
                
                print(f"✅ Plugin execution completed:")
                print(f"   Duration: {result.duration:.3f} seconds")
                if result.error:
                    print(f"   Error: {result.error}")
                else:
                    print(f"   Result: {result.result}")
            else:
                print("🔌 No plugins available for execution")
        except Exception as e:
            print(f"⚠️  Plugin execution failed: {e}")


if __name__ == "__main__":
    try:
        main()
        print("\n✅ Example completed successfully!")
    except gzhclient.GZHConnectionError as e:
        print(f"\n❌ Connection error: {e}")
        print("💡 Make sure the GZH Manager library is built and available.")
        print("💡 Build the library with: cd bindings/python && go build -buildmode=c-shared -o libgzh.so libgzh.go")
        sys.exit(1)
    except gzhclient.GZHError as e:
        print(f"\n❌ GZH Manager error: {e}")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\n⚠️  Interrupted by user")
        sys.exit(0)
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")
        sys.exit(1)