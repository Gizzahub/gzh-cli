// Package event implements an event-driven architecture for the gzh-manager-go project.
// It provides a publish-subscribe mechanism for decoupling components and enabling
// asynchronous communication between different parts of the system.
//
// The package includes:
//   - Event bus for publishing and subscribing to events
//   - Event types for different system activities
//   - Event handlers with filtering and routing
//   - Webhook event processing for Git services
//   - Event persistence and replay capabilities
//   - Metrics and monitoring integration
//
// Common event types:
//   - Repository cloned/updated/deleted
//   - Configuration changed
//   - Network environment switched
//   - Build/test completed
//   - Error occurred
package event