"""
GZH Manager Python Client Exceptions

Custom exception classes for the GZH Manager Python client.
"""


class GZHError(Exception):
    """Base exception class for GZH Manager errors."""
    
    def __init__(self, message, code=None, details=None):
        super().__init__(message)
        self.message = message
        self.code = code
        self.details = details or {}
    
    def __str__(self):
        if self.code:
            return f"[{self.code}] {self.message}"
        return self.message


class GZHConnectionError(GZHError):
    """Raised when connection to GZH Manager fails."""
    pass


class GZHAPIError(GZHError):
    """Raised when GZH Manager API returns an error."""
    
    def __init__(self, message, code=None, details=None, status_code=None):
        super().__init__(message, code, details)
        self.status_code = status_code


class GZHConfigurationError(GZHError):
    """Raised when GZH Manager configuration is invalid."""
    pass


class GZHPluginError(GZHError):
    """Raised when plugin operations fail."""
    pass


class GZHTimeoutError(GZHError):
    """Raised when operations timeout."""
    pass