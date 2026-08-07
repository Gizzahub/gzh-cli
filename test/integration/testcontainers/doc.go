// Package testcontainers previously held stub container helpers that never
// started real Docker containers (issue 43). Real testcontainers support is
// deferred; do not reintroduce no-op Setup* helpers that pretend to integrate
// with forges.
//
// Integration coverage that only needs config I/O lives under
// test/integration/docker without this package.
package testcontainers
