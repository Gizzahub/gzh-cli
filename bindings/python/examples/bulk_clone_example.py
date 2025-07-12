#!/usr/bin/env python3
"""
GZH Manager Python Client - Bulk Clone Example

This example demonstrates advanced bulk cloning features.
"""

import os
import sys
from pathlib import Path
from datetime import datetime, timedelta

# Add the parent directory to Python path so we can import gzhclient
sys.path.insert(0, str(Path(__file__).parent.parent))

import gzhclient
from gzhclient import PlatformConfig, CloneFilters, ClientConfig


def setup_platforms():
    """Setup platform configurations."""
    platforms = []
    
    # GitHub configuration
    github_token = os.getenv("GITHUB_TOKEN")
    if github_token:
        platforms.append(PlatformConfig(
            type="github",
            token=github_token,
            organizations=["golang", "kubernetes"],  # Example organizations
            users=["torvalds"]  # Example user
        ))
        print("✅ GitHub platform configured")
    else:
        print("⚠️  GITHUB_TOKEN not set, skipping GitHub")
    
    # GitLab configuration
    gitlab_token = os.getenv("GITLAB_TOKEN")
    if gitlab_token:
        platforms.append(PlatformConfig(
            type="gitlab",
            url="https://gitlab.com",
            token=gitlab_token,
            organizations=["gitlab-org"]  # Example organization
        ))
        print("✅ GitLab platform configured")
    else:
        print("⚠️  GITLAB_TOKEN not set, skipping GitLab")
    
    # Gitea configuration (if available)
    gitea_token = os.getenv("GITEA_TOKEN")
    gitea_url = os.getenv("GITEA_URL", "https://gitea.com")
    if gitea_token:
        platforms.append(PlatformConfig(
            type="gitea",
            url=gitea_url,
            token=gitea_token,
            organizations=["gitea"]  # Example organization
        ))
        print("✅ Gitea platform configured")
    else:
        print("⚠️  GITEA_TOKEN not set, skipping Gitea")
    
    return platforms


def setup_filters():
    """Setup clone filters."""
    return CloneFilters(
        # Only include repositories with these languages
        languages=["go", "python", "javascript", "typescript"],
        
        # Exclude test and archived repositories
        exclude_repos=[
            "test-*",
            "*-test",
            "archive-*",
            "deprecated-*",
            "old-*"
        ],
        
        # Only repositories updated in the last 6 months
        updated_after=datetime.now() - timedelta(days=180),
        
        # Size constraints (in bytes)
        min_size=1024,  # At least 1KB
        max_size=100 * 1024 * 1024,  # At most 100MB
    )


def perform_bulk_clone(client, platforms, output_dir, dry_run=True):
    """Perform bulk clone operation."""
    
    if not platforms:
        print("❌ No platforms configured. Please set tokens in environment variables.")
        return
    
    filters = setup_filters()
    
    print(f"\n📋 Bulk Clone Configuration:")
    print(f"   Platforms: {len(platforms)}")
    print(f"   Output directory: {output_dir}")
    print(f"   Concurrency: 3")
    print(f"   Strategy: reset")
    print(f"   Include private: False")
    print(f"   Language filters: {', '.join(filters.languages)}")
    print(f"   Exclude patterns: {', '.join(filters.exclude_repos)}")
    print(f"   Updated after: {filters.updated_after.strftime('%Y-%m-%d')}")
    print(f"   Size range: {filters.min_size} - {filters.max_size} bytes")
    
    if dry_run:
        print("\n🔍 DRY RUN MODE - No repositories will be cloned")
        print("💡 Set DRY_RUN=false to perform actual cloning")
        return
    
    print(f"\n🚀 Starting bulk clone operation...")
    
    try:
        result = client.bulk_clone(
            platforms=platforms,
            output_dir=output_dir,
            concurrency=3,  # Conservative concurrency for example
            strategy="reset",
            include_private=False,
            filters=filters
        )
        
        print(f"\n📊 Bulk Clone Results:")
        print(f"   Total repositories: {result.total_repos}")
        print(f"   Successfully cloned: {result.success_count}")
        print(f"   Failed: {result.failure_count}")
        print(f"   Skipped: {result.skipped_count}")
        print(f"   Duration: {result.duration:.2f} seconds")
        
        # Show summary statistics
        if result.summary:
            print(f"\n📈 Summary Statistics:")
            for key, value in result.summary.items():
                print(f"   {key}: {value}")
        
        # Show successful clones
        if result.success_count > 0:
            print(f"\n✅ Successfully Cloned Repositories:")
            success_repos = [r for r in result.results if r.status == "success"]
            for repo in success_repos[:10]:  # Show first 10
                size_mb = repo.size / (1024 * 1024) if repo.size > 0 else 0
                print(f"   📁 {repo.repo_name} ({repo.platform}) - {size_mb:.1f}MB")
            
            if len(success_repos) > 10:
                print(f"   ... and {len(success_repos) - 10} more")
        
        # Show failures
        if result.failure_count > 0:
            print(f"\n❌ Failed Repositories:")
            failed_repos = [r for r in result.results if r.status == "failed"]
            for repo in failed_repos:
                print(f"   ❌ {repo.repo_name} ({repo.platform}): {repo.error}")
        
        # Show skipped repositories
        if result.skipped_count > 0:
            print(f"\n⏭️  Skipped Repositories:")
            skipped_repos = [r for r in result.results if r.status == "skipped"]
            for repo in skipped_repos[:5]:  # Show first 5
                print(f"   ⏭️  {repo.repo_name} ({repo.platform})")
            
            if len(skipped_repos) > 5:
                print(f"   ... and {len(skipped_repos) - 5} more")
        
    except gzhclient.GZHTimeoutError as e:
        print(f"⏰ Operation timed out: {e}")
    except gzhclient.GZHAPIError as e:
        print(f"💥 API error: {e}")
    except Exception as e:
        print(f"💥 Unexpected error: {e}")


def main():
    """Main example function."""
    
    print("🚀 GZH Manager Python Client - Bulk Clone Example")
    print("=" * 55)
    
    # Configuration
    config = ClientConfig(
        timeout=300,  # 5 minutes timeout
        retry_count=3,
        enable_plugins=False,  # Disable plugins for this example
        log_level="info"
    )
    
    # Setup
    platforms = setup_platforms()
    output_dir = Path("./example_repositories")
    dry_run = os.getenv("DRY_RUN", "true").lower() != "false"
    
    # Create client and perform bulk clone
    print(f"\n🔧 Creating client with {config.timeout}s timeout...")
    
    with gzhclient.Client(config) as client:
        print("✅ Client created successfully")
        
        # Check client health before proceeding
        health = client.health()
        if health.overall != gzhclient.StatusType.HEALTHY:
            print(f"⚠️  Client health issue: {health.overall.value}")
            for name, component in health.components.items():
                if component.status != gzhclient.StatusType.HEALTHY:
                    print(f"   ⚠️  {name}: {component.message}")
        
        # Perform bulk clone
        perform_bulk_clone(client, platforms, output_dir, dry_run)


if __name__ == "__main__":
    try:
        main()
        print("\n✅ Bulk clone example completed!")
    except gzhclient.GZHConnectionError as e:
        print(f"\n❌ Connection error: {e}")
        print("💡 Make sure the GZH Manager library is built and available.")
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