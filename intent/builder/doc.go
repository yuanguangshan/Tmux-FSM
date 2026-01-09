// Package builder provides NATIVE Intent builders.
//
// This package is the ONLY authoritative way to construct new Intents.
// Legacy intent construction paths are frozen elsewhere and must not be extended.
//
// Rules:
// - Do NOT import legacy logic
// - Builders must be semantic-only
// - Priority determines builder matching order
package builder
