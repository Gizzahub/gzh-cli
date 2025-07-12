"""
GZH Manager Python Client

A Python client library for GZH Manager, providing programmatic access to
bulk repository operations, plugin management, system monitoring, and more.

Example:
    >>> import gzhclient
    >>> client = gzhclient.Client()
    >>> result = client.bulk_clone([
    ...     {
    ...         'type': 'github',
    ...         'token': 'your-token',
    ...         'organizations': ['your-org']
    ...     }
    ... ], output_dir='./repos')
    >>> print(f"Cloned {result['success_count']} repositories")
"""

from .client import Client, ClientConfig
from .exceptions import GZHError, GZHConnectionError, GZHAPIError
from .types import (
    BulkCloneRequest, BulkCloneResult, RepositoryCloneResult,
    PlatformConfig, CloneFilters, PluginInfo, PluginExecuteRequest,
    PluginExecuteResult, SystemMetrics, HealthStatus
)

__version__ = "1.0.0"
__author__ = "GZH Manager Team"
__email__ = "support@gzh-manager.com"

__all__ = [
    "Client",
    "ClientConfig", 
    "GZHError",
    "GZHConnectionError",
    "GZHAPIError",
    "BulkCloneRequest",
    "BulkCloneResult", 
    "RepositoryCloneResult",
    "PlatformConfig",
    "CloneFilters",
    "PluginInfo",
    "PluginExecuteRequest",
    "PluginExecuteResult",
    "SystemMetrics",
    "HealthStatus"
]