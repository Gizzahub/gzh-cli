// Package serve implements the HTTP server functionality for the GZH Manager
// web interface and API endpoints.
//
// This package provides the serve command that starts a web server hosting
// the GZH Manager dashboard and RESTful API, enabling web-based management
// of GZH operations and real-time monitoring through a browser interface.
//
// Key Features:
//
// Web Dashboard:
//   - Real-time monitoring interface
//   - Interactive configuration management
//   - Operation status and progress tracking
//   - System health and metrics visualization
//   - Responsive design for mobile and desktop
//
// RESTful API:
//   - Repository management endpoints
//   - Configuration CRUD operations
//   - Monitoring and metrics API
//   - Webhook handling and processing
//   - Authentication and authorization
//
// Static File Serving:
//   - React application hosting
//   - Asset optimization and caching
//   - Development and production modes
//   - Custom static content support
//
// Security Features:
//   - HTTPS/TLS support
//   - CORS configuration
//   - Rate limiting and throttling
//   - Request validation and sanitization
//   - Secure headers and CSP
//
// Example usage:
//
//	gz serve --port 8080
//	gz serve --port 8080 --static-dir web/build
//	gz serve --https --cert cert.pem --key key.pem
//	gz serve --api-only --port 3000
//
// The server supports both development and production deployments,
// with automatic detection of the React build directory and
// intelligent routing between API endpoints and static content.
package serve
