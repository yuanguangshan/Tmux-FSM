# Replay vs Verifier

## The Naive Assumption

> "If the final text is correct, the history must be correct."

This assumption is false.

---

## Comparison

| Dimension | Replay Engine | Verifier |
|--------|--------------|----------|
| Trusts input order | ✅ Yes | ❌ No |
| Detects parent mismatch | ❌ No | ✅ Yes |
| Detects reordered history | ❌ No | ✅ Yes |
| Commits to full history | ❌ No | ✅ Yes |
| Same output, different history | ❌ Undetectable | ✅ Different roots |
| Deterministic verification | ❌ Engine-dependent | ✅ Protocol-defined |

---

## Visual Example

### History A
```
H1 ──▶ H2
 A     B
```

### History B
```
H1'
 AB
```

Both replay to:

```
"AB"
```

But verifier computes:

```
StateRoot(A) ≠ StateRoot(B)
```

---

## Why This Matters

Replay answers:
> "Does this run?"

Verifier answers:
> "Was this the *only* possible history?"

Only the verifier enables:
- Auditing
- Fork detection
- Trustless replication
- Cryptographic commitments