"""
GZH Manager Python Client Types

Type definitions and data classes for the GZH Manager Python client.
"""

from dataclasses import dataclass, field
from typing import List, Dict, Any, Optional, Union
from datetime import datetime
from enum import Enum


class StatusType(Enum):
    """Health status enumeration."""
    HEALTHY = "healthy"
    DEGRADED = "degraded"
    UNHEALTHY = "unhealthy"


@dataclass
class ClientConfig:
    """Configuration for GZH Manager client."""
    timeout: int = 30  # seconds
    retry_count: int = 3
    enable_plugins: bool = True
    plugin_dir: Optional[str] = None
    log_level: str = "info"
    log_file: Optional[str] = None
    features: Dict[str, bool] = field(default_factory=lambda: {
        "bulk_clone": True,
        "dev_env": True,
        "net_env": True,
        "monitoring": True,
        "plugins": True
    })


@dataclass
class PlatformConfig:
    """Configuration for a Git platform."""
    type: str  # github, gitlab, gitea, gogs
    url: Optional[str] = None
    token: Optional[str] = None
    organizations: List[str] = field(default_factory=list)
    users: List[str] = field(default_factory=list)


@dataclass
class CloneFilters:
    """Filtering options for repository cloning."""
    include_repos: List[str] = field(default_factory=list)
    exclude_repos: List[str] = field(default_factory=list)
    languages: List[str] = field(default_factory=list)
    min_size: Optional[int] = None
    max_size: Optional[int] = None
    updated_after: Optional[datetime] = None


@dataclass
class BulkCloneRequest:
    """Request for bulk repository cloning operation."""
    platforms: List[PlatformConfig]
    output_dir: str
    concurrency: int = 5
    strategy: str = "reset"  # reset, pull, fetch
    include_private: bool = False
    filters: CloneFilters = field(default_factory=CloneFilters)


@dataclass
class RepositoryCloneResult:
    """Result of cloning a single repository."""
    repo_name: str
    platform: str
    url: str
    local_path: str
    status: str  # success, failed, skipped
    error: Optional[str] = None
    duration: float = 0.0  # seconds
    size: int = 0  # bytes


@dataclass
class BulkCloneResult:
    """Result of bulk repository cloning operation."""
    total_repos: int
    success_count: int
    failure_count: int
    skipped_count: int
    results: List[RepositoryCloneResult]
    duration: float  # seconds
    summary: Dict[str, Any] = field(default_factory=dict)


@dataclass
class PluginInfo:
    """Information about a plugin."""
    name: str
    version: str
    description: str
    author: str
    status: str
    capabilities: List[str]
    load_time: datetime
    last_used: datetime
    call_count: int
    error_count: int


@dataclass
class PluginExecuteRequest:
    """Request for plugin execution."""
    plugin_name: str
    method: Optional[str] = None
    args: Dict[str, Any] = field(default_factory=dict)
    timeout: int = 30  # seconds


@dataclass
class PluginExecuteResult:
    """Result of plugin execution."""
    plugin_name: str
    method: str
    result: Any
    error: Optional[str] = None
    duration: float  # seconds
    timestamp: datetime = field(default_factory=datetime.now)


@dataclass
class CPUMetrics:
    """CPU metrics."""
    usage: float = 0.0
    cores: int = 0
    user_time: float = 0.0
    system_time: float = 0.0
    idle_time: float = 0.0


@dataclass
class MemoryMetrics:
    """Memory metrics."""
    total: int = 0
    used: int = 0
    available: int = 0
    usage: float = 0.0
    cached: int = 0
    buffers: int = 0


@dataclass
class DiskMetrics:
    """Disk metrics."""
    total: int = 0
    used: int = 0
    available: int = 0
    usage: float = 0.0
    read_ops: int = 0
    write_ops: int = 0
    read_bytes: int = 0
    write_bytes: int = 0


@dataclass
class SystemMetrics:
    """System-level metrics."""
    cpu: CPUMetrics = field(default_factory=CPUMetrics)
    memory: MemoryMetrics = field(default_factory=MemoryMetrics)
    disk: DiskMetrics = field(default_factory=DiskMetrics)
    load_avg: List[float] = field(default_factory=list)
    uptime: float = 0.0  # seconds
    timestamp: datetime = field(default_factory=datetime.now)


@dataclass
class ComponentHealth:
    """Health status of a component."""
    status: StatusType
    message: str
    details: Dict[str, Any] = field(default_factory=dict)


@dataclass
class HealthStatus:
    """Overall health status."""
    overall: StatusType
    components: Dict[str, ComponentHealth]
    timestamp: datetime = field(default_factory=datetime.now)