// Package audit provides audit logging and compliance tracking
// functionality for the GZH Manager system.
//
// This package implements comprehensive audit trails, compliance
// monitoring, and security logging to track system activities,
// changes, and access patterns for security and regulatory compliance.
//
// Key Components:
//
// Audit Logger:
//   - Structured audit event logging
//   - Tamper-evident log storage
//   - Event correlation and tracking
//   - Log integrity verification
//
// Activity Tracking:
//   - User action monitoring
//   - System change tracking
//   - Access pattern analysis
//   - Administrative operation logging
//
// Compliance Engine:
//   - Regulatory compliance monitoring
//   - Policy violation detection
//   - Compliance report generation
//   - Audit trail validation
//
// Security Monitoring:
//   - Suspicious activity detection
//   - Security event correlation
//   - Threat intelligence integration
//   - Real-time security alerting
//
// Features:
//   - Real-time audit logging
//   - Long-term audit retention
//   - Searchable audit trails
//   - Automated compliance reporting
//   - Integration with SIEM systems
//
// Example usage:
//
//	auditor := audit.NewAuditor(config)
//
//	event := audit.NewEvent("user.login", userID)
//	event.AddDetail("source_ip", clientIP)
//	auditor.LogEvent(event)
//
//	report := auditor.GenerateComplianceReport(period)
//	violations := auditor.CheckCompliance(policy)
//
// The package ensures comprehensive audit trails and compliance
// monitoring for security and regulatory requirements.
package audit
