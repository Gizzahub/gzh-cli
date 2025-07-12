"""
GZH Manager Python Client

Main client class for interacting with GZH Manager functionality.
"""

import json
import ctypes
import platform
import os
from typing import List, Dict, Any, Optional
from datetime import datetime, timezone
from pathlib import Path

from .types import (
    ClientConfig, BulkCloneRequest, BulkCloneResult, RepositoryCloneResult,
    PlatformConfig, CloneFilters, PluginInfo, PluginExecuteRequest,
    PluginExecuteResult, SystemMetrics, HealthStatus, ComponentHealth,
    StatusType, CPUMetrics, MemoryMetrics, DiskMetrics
)
from .exceptions import (
    GZHError, GZHConnectionError, GZHAPIError, GZHConfigurationError,
    GZHPluginError, GZHTimeoutError
)


class _CClientConfig(ctypes.Structure):
    """C structure for client configuration."""
    _fields_ = [
        ("timeout", ctypes.c_int64),
        ("retry_count", ctypes.c_int),
        ("enable_plugins", ctypes.c_int),
        ("plugin_dir", ctypes.c_char_p),
        ("log_level", ctypes.c_char_p),
        ("log_file", ctypes.c_char_p),
    ]


class _CBulkCloneRequest(ctypes.Structure):
    """C structure for bulk clone request."""
    _fields_ = [
        ("platforms_json", ctypes.c_char_p),
        ("output_dir", ctypes.c_char_p),
        ("concurrency", ctypes.c_int),
        ("strategy", ctypes.c_char_p),
        ("include_private", ctypes.c_int),
        ("filters_json", ctypes.c_char_p),
    ]


class _CResult(ctypes.Structure):
    """C structure for operation result."""
    _fields_ = [
        ("success", ctypes.c_int),
        ("error_msg", ctypes.c_char_p),
        ("data_json", ctypes.c_char_p),
    ]


class Client:
    """
    GZH Manager Python Client
    
    Provides programmatic access to GZH Manager functionality including
    bulk repository operations, plugin management, and system monitoring.
    
    Example:
        >>> client = Client()
        >>> health = client.health()
        >>> print(f"Client status: {health.overall.value}")
        
        >>> result = client.bulk_clone([
        ...     PlatformConfig(
        ...         type='github',
        ...         token='your-token',
        ...         organizations=['your-org']
        ...     )
        ... ], output_dir='./repos')
        >>> print(f"Cloned {result.success_count} repositories")
    """
    
    def __init__(self, config: Optional[ClientConfig] = None):
        """
        Initialize GZH Manager client.
        
        Args:
            config: Client configuration. If None, uses default configuration.
            
        Raises:
            GZHConnectionError: If unable to connect to GZH Manager.
            GZHConfigurationError: If configuration is invalid.
        """
        self._lib = None
        self._client_id = None
        
        try:
            self._load_library()
            self._client_id = self._create_client(config or ClientConfig())
        except Exception as e:
            raise GZHConnectionError(f"Failed to initialize client: {str(e)}")
    
    def __enter__(self):
        """Context manager entry."""
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()
    
    def close(self):
        """Close the client and cleanup resources."""
        if self._lib and self._client_id is not None:
            self._lib.gzh_client_destroy(self._client_id)
            self._client_id = None
    
    def health(self) -> HealthStatus:
        """
        Get client health status.
        
        Returns:
            HealthStatus: Current health status of the client.
            
        Raises:
            GZHAPIError: If health check fails.
        """
        result = self._lib.gzh_client_health(self._client_id)
        try:
            data = self._process_result(result)
            return self._parse_health_status(data)
        finally:
            self._lib.gzh_free_result(result)
    
    def bulk_clone(
        self,
        platforms: List[PlatformConfig],
        output_dir: str,
        concurrency: int = 5,
        strategy: str = "reset",
        include_private: bool = False,
        filters: Optional[CloneFilters] = None
    ) -> BulkCloneResult:
        """
        Perform bulk repository cloning operation.
        
        Args:
            platforms: List of platform configurations.
            output_dir: Directory to clone repositories into.
            concurrency: Number of concurrent clone operations.
            strategy: Clone strategy ("reset", "pull", "fetch").
            include_private: Whether to include private repositories.
            filters: Optional filtering criteria.
            
        Returns:
            BulkCloneResult: Result of the bulk clone operation.
            
        Raises:
            GZHAPIError: If bulk clone operation fails.
        """
        request = BulkCloneRequest(
            platforms=platforms,
            output_dir=output_dir,
            concurrency=concurrency,
            strategy=strategy,
            include_private=include_private,
            filters=filters or CloneFilters()
        )
        
        c_request = self._prepare_bulk_clone_request(request)
        result = self._lib.gzh_bulk_clone(self._client_id, ctypes.byref(c_request))
        
        try:
            data = self._process_result(result)
            return self._parse_bulk_clone_result(data)
        finally:
            self._lib.gzh_free_result(result)
    
    def list_plugins(self) -> List[PluginInfo]:
        """
        List available plugins.
        
        Returns:
            List[PluginInfo]: List of available plugins.
            
        Raises:
            GZHPluginError: If unable to list plugins.
        """
        result = self._lib.gzh_list_plugins(self._client_id)
        try:
            data = self._process_result(result)
            return self._parse_plugin_list(data)
        except GZHError as e:
            raise GZHPluginError(str(e))
        finally:
            self._lib.gzh_free_result(result)
    
    def execute_plugin(
        self,
        plugin_name: str,
        method: Optional[str] = None,
        args: Optional[Dict[str, Any]] = None,
        timeout: int = 30
    ) -> PluginExecuteResult:
        """
        Execute a plugin method.
        
        Args:
            plugin_name: Name of the plugin to execute.
            method: Method to call (optional).
            args: Arguments to pass to the plugin.
            timeout: Execution timeout in seconds.
            
        Returns:
            PluginExecuteResult: Result of plugin execution.
            
        Raises:
            GZHPluginError: If plugin execution fails.
        """
        args_json = json.dumps(args or {}).encode('utf-8')
        method_str = (method or "").encode('utf-8')
        plugin_name_str = plugin_name.encode('utf-8')
        
        result = self._lib.gzh_execute_plugin(
            self._client_id,
            plugin_name_str,
            method_str,
            args_json,
            timeout
        )
        
        try:
            data = self._process_result(result)
            return self._parse_plugin_execute_result(data)
        except GZHError as e:
            raise GZHPluginError(str(e))
        finally:
            self._lib.gzh_free_result(result)
    
    def get_system_metrics(self) -> SystemMetrics:
        """
        Get current system metrics.
        
        Returns:
            SystemMetrics: Current system metrics.
            
        Raises:
            GZHAPIError: If unable to get system metrics.
        """
        result = self._lib.gzh_get_system_metrics(self._client_id)
        try:
            data = self._process_result(result)
            return self._parse_system_metrics(data)
        finally:
            self._lib.gzh_free_result(result)
    
    def _load_library(self):
        """Load the GZH Manager shared library."""
        # Determine library name based on platform
        system = platform.system().lower()
        if system == "windows":
            lib_name = "libgzh.dll"
        elif system == "darwin":
            lib_name = "libgzh.dylib"
        else:
            lib_name = "libgzh.so"
        
        # Try to find library in various locations
        lib_paths = [
            # Current directory
            Path(__file__).parent / lib_name,
            # Parent directory
            Path(__file__).parent.parent / lib_name,
            # System library paths
            Path("/usr/local/lib") / lib_name,
            Path("/usr/lib") / lib_name,
        ]
        
        # Add GZHCLIENT_LIB_PATH environment variable
        if "GZHCLIENT_LIB_PATH" in os.environ:
            lib_paths.insert(0, Path(os.environ["GZHCLIENT_LIB_PATH"]) / lib_name)
        
        for lib_path in lib_paths:
            if lib_path.exists():
                try:
                    self._lib = ctypes.CDLL(str(lib_path))
                    break
                except OSError:
                    continue
        
        if not self._lib:
            raise GZHConnectionError(
                f"Could not load GZH Manager library ({lib_name}). "
                f"Please ensure it is built and available in one of: {lib_paths}"
            )
        
        # Set up function signatures
        self._setup_function_signatures()
    
    def _setup_function_signatures(self):
        """Set up C function signatures."""
        # gzh_client_create
        self._lib.gzh_client_create.argtypes = [ctypes.POINTER(_CClientConfig)]
        self._lib.gzh_client_create.restype = ctypes.c_int
        
        # gzh_client_destroy
        self._lib.gzh_client_destroy.argtypes = [ctypes.c_int]
        self._lib.gzh_client_destroy.restype = None
        
        # gzh_client_health
        self._lib.gzh_client_health.argtypes = [ctypes.c_int]
        self._lib.gzh_client_health.restype = ctypes.POINTER(_CResult)
        
        # gzh_bulk_clone
        self._lib.gzh_bulk_clone.argtypes = [ctypes.c_int, ctypes.POINTER(_CBulkCloneRequest)]
        self._lib.gzh_bulk_clone.restype = ctypes.POINTER(_CResult)
        
        # gzh_list_plugins
        self._lib.gzh_list_plugins.argtypes = [ctypes.c_int]
        self._lib.gzh_list_plugins.restype = ctypes.POINTER(_CResult)
        
        # gzh_execute_plugin
        self._lib.gzh_execute_plugin.argtypes = [
            ctypes.c_int, ctypes.c_char_p, ctypes.c_char_p, 
            ctypes.c_char_p, ctypes.c_int
        ]
        self._lib.gzh_execute_plugin.restype = ctypes.POINTER(_CResult)
        
        # gzh_get_system_metrics
        self._lib.gzh_get_system_metrics.argtypes = [ctypes.c_int]
        self._lib.gzh_get_system_metrics.restype = ctypes.POINTER(_CResult)
        
        # gzh_free_result
        self._lib.gzh_free_result.argtypes = [ctypes.POINTER(_CResult)]
        self._lib.gzh_free_result.restype = None
    
    def _create_client(self, config: ClientConfig) -> int:
        """Create a new client instance."""
        c_config = _CClientConfig()
        c_config.timeout = config.timeout
        c_config.retry_count = config.retry_count
        c_config.enable_plugins = 1 if config.enable_plugins else 0
        
        if config.plugin_dir:
            c_config.plugin_dir = config.plugin_dir.encode('utf-8')
        if config.log_level:
            c_config.log_level = config.log_level.encode('utf-8')
        if config.log_file:
            c_config.log_file = config.log_file.encode('utf-8')
        
        client_id = self._lib.gzh_client_create(ctypes.byref(c_config))
        if client_id < 0:
            raise GZHConnectionError("Failed to create client")
        
        return client_id
    
    def _prepare_bulk_clone_request(self, request: BulkCloneRequest) -> _CBulkCloneRequest:
        """Prepare C bulk clone request structure."""
        platforms_data = []
        for platform in request.platforms:
            platform_dict = {
                "type": platform.type,
                "url": platform.url,
                "token": platform.token,
                "organizations": platform.organizations,
                "users": platform.users
            }
            platforms_data.append(platform_dict)
        
        filters_dict = {
            "include_repos": request.filters.include_repos,
            "exclude_repos": request.filters.exclude_repos,
            "languages": request.filters.languages,
            "min_size": request.filters.min_size,
            "max_size": request.filters.max_size,
            "updated_after": request.filters.updated_after.isoformat() if request.filters.updated_after else None
        }
        
        c_request = _CBulkCloneRequest()
        c_request.platforms_json = json.dumps(platforms_data).encode('utf-8')
        c_request.output_dir = request.output_dir.encode('utf-8')
        c_request.concurrency = request.concurrency
        c_request.strategy = request.strategy.encode('utf-8')
        c_request.include_private = 1 if request.include_private else 0
        c_request.filters_json = json.dumps(filters_dict).encode('utf-8')
        
        return c_request
    
    def _process_result(self, result: ctypes.POINTER(_CResult)) -> Dict[str, Any]:
        """Process C result and handle errors."""
        if not result:
            raise GZHAPIError("Null result returned")
        
        if result.contents.success == 0:
            error_msg = "Unknown error"
            if result.contents.error_msg:
                error_msg = result.contents.error_msg.decode('utf-8')
            raise GZHAPIError(error_msg)
        
        if not result.contents.data_json:
            raise GZHAPIError("No data returned")
        
        data_str = result.contents.data_json.decode('utf-8')
        try:
            return json.loads(data_str)
        except json.JSONDecodeError as e:
            raise GZHAPIError(f"Invalid JSON response: {str(e)}")
    
    def _parse_health_status(self, data: Dict[str, Any]) -> HealthStatus:
        """Parse health status from JSON data."""
        overall = StatusType(data.get("overall", "unknown"))
        
        components = {}
        for name, comp_data in data.get("components", {}).items():
            components[name] = ComponentHealth(
                status=StatusType(comp_data.get("status", "unknown")),
                message=comp_data.get("message", ""),
                details=comp_data.get("details", {})
            )
        
        timestamp = datetime.fromisoformat(data.get("timestamp", datetime.now().isoformat()))
        
        return HealthStatus(
            overall=overall,
            components=components,
            timestamp=timestamp
        )
    
    def _parse_bulk_clone_result(self, data: Dict[str, Any]) -> BulkCloneResult:
        """Parse bulk clone result from JSON data."""
        results = []
        for result_data in data.get("results", []):
            results.append(RepositoryCloneResult(
                repo_name=result_data.get("repo_name", ""),
                platform=result_data.get("platform", ""),
                url=result_data.get("url", ""),
                local_path=result_data.get("local_path", ""),
                status=result_data.get("status", ""),
                error=result_data.get("error"),
                duration=result_data.get("duration", 0.0) / 1e9,  # Convert nanoseconds to seconds
                size=result_data.get("size", 0)
            ))
        
        return BulkCloneResult(
            total_repos=data.get("total_repos", 0),
            success_count=data.get("success_count", 0),
            failure_count=data.get("failure_count", 0),
            skipped_count=data.get("skipped_count", 0),
            results=results,
            duration=data.get("duration", 0.0) / 1e9,  # Convert nanoseconds to seconds
            summary=data.get("summary", {})
        )
    
    def _parse_plugin_list(self, data: List[Dict[str, Any]]) -> List[PluginInfo]:
        """Parse plugin list from JSON data."""
        plugins = []
        for plugin_data in data:
            load_time = datetime.fromisoformat(plugin_data.get("load_time", datetime.now().isoformat()))
            last_used = datetime.fromisoformat(plugin_data.get("last_used", datetime.now().isoformat()))
            
            plugins.append(PluginInfo(
                name=plugin_data.get("name", ""),
                version=plugin_data.get("version", ""),
                description=plugin_data.get("description", ""),
                author=plugin_data.get("author", ""),
                status=plugin_data.get("status", ""),
                capabilities=plugin_data.get("capabilities", []),
                load_time=load_time,
                last_used=last_used,
                call_count=plugin_data.get("call_count", 0),
                error_count=plugin_data.get("error_count", 0)
            ))
        
        return plugins
    
    def _parse_plugin_execute_result(self, data: Dict[str, Any]) -> PluginExecuteResult:
        """Parse plugin execute result from JSON data."""
        timestamp = datetime.fromisoformat(data.get("timestamp", datetime.now().isoformat()))
        
        return PluginExecuteResult(
            plugin_name=data.get("plugin_name", ""),
            method=data.get("method", ""),
            result=data.get("result"),
            error=data.get("error"),
            duration=data.get("duration", 0.0) / 1e9,  # Convert nanoseconds to seconds
            timestamp=timestamp
        )
    
    def _parse_system_metrics(self, data: Dict[str, Any]) -> SystemMetrics:
        """Parse system metrics from JSON data."""
        cpu_data = data.get("cpu", {})
        memory_data = data.get("memory", {})
        disk_data = data.get("disk", {})
        
        timestamp = datetime.fromisoformat(data.get("timestamp", datetime.now().isoformat()))
        
        return SystemMetrics(
            cpu=CPUMetrics(
                usage=cpu_data.get("usage", 0.0),
                cores=cpu_data.get("cores", 0),
                user_time=cpu_data.get("user_time", 0.0),
                system_time=cpu_data.get("system_time", 0.0),
                idle_time=cpu_data.get("idle_time", 0.0)
            ),
            memory=MemoryMetrics(
                total=memory_data.get("total", 0),
                used=memory_data.get("used", 0),
                available=memory_data.get("available", 0),
                usage=memory_data.get("usage", 0.0),
                cached=memory_data.get("cached", 0),
                buffers=memory_data.get("buffers", 0)
            ),
            disk=DiskMetrics(
                total=disk_data.get("total", 0),
                used=disk_data.get("used", 0),
                available=disk_data.get("available", 0),
                usage=disk_data.get("usage", 0.0),
                read_ops=disk_data.get("read_ops", 0),
                write_ops=disk_data.get("write_ops", 0),
                read_bytes=disk_data.get("read_bytes", 0),
                write_bytes=disk_data.get("write_bytes", 0)
            ),
            load_avg=data.get("load_avg", []),
            uptime=data.get("uptime", 0.0) / 1e9,  # Convert nanoseconds to seconds
            timestamp=timestamp
        )