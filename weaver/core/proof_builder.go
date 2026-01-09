package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// ProofBuilder builds proof objects for audit-compliant transactions
type ProofBuilder struct{}

// NewProofBuilder creates a new ProofBuilder instance
func NewProofBuilder() *ProofBuilder {
	return &ProofBuilder{}
}

// BuildProof creates a proof object from transaction data
func (pb *ProofBuilder) BuildProof(tx *Transaction, auditRecord *AuditRecord) *Proof {
	if tx == nil {
		return nil
	}

	// Calculate hashes for the proof
	preStateHash := pb.calculateHash(tx.Intent.GetSnapshotHash()) // Using the original snapshot hash as pre-state
	postStateHash := pb.calculateHash(tx.PostSnapshotHash)
	factsHash := pb.calculateFactsHash(tx.Facts)
	auditHash := pb.calculateAuditHash(auditRecord)

	return &Proof{
		TransactionID: string(tx.ID),
		PreStateHash:  preStateHash,
		PostStateHash: postStateHash,
		FactsHash:     factsHash,
		AuditHash:     auditHash,
	}
}

// calculateHash creates a SHA256 hash of the input string
func (pb *ProofBuilder) calculateHash(input string) string {
	if input == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// calculateFactsHash creates a hash of the facts array
func (pb *ProofBuilder) calculateFactsHash(facts []Fact) string {
	if len(facts) == 0 {
		return ""
	}
	
	// Serialize facts to JSON for consistent hashing
	factsJSON, err := json.Marshal(facts)
	if err != nil {
		return ""
	}
	
	hash := sha256.Sum256(factsJSON)
	return hex.EncodeToString(hash[:])
}

// calculateAuditHash creates a hash of the audit record
func (pb *ProofBuilder) calculateAuditHash(auditRecord *AuditRecord) string {
	if auditRecord == nil {
		return ""
	}
	
	// Serialize audit record to JSON for consistent hashing
	auditJSON, err := json.Marshal(auditRecord)
	if err != nil {
		return ""
	}
	
	hash := sha256.Sum256(auditJSON)
	return hex.EncodeToString(hash[:])
}

// VerifyProof checks if the proof is valid by recomputing hashes
func (pb *ProofBuilder) VerifyProof(proof *Proof, tx *Transaction, auditRecord *AuditRecord) bool {
	if proof == nil || tx == nil {
		return false
	}

	// Recompute the proof
	recomputedProof := pb.BuildProof(tx, auditRecord)
	if recomputedProof == nil {
		return false
	}

	// Compare all hashes
	return proof.TransactionID == recomputedProof.TransactionID &&
		proof.PreStateHash == recomputedProof.PreStateHash &&
		proof.PostStateHash == recomputedProof.PostStateHash &&
		proof.FactsHash == recomputedProof.FactsHash &&
		proof.AuditHash == recomputedProof.AuditHash
}