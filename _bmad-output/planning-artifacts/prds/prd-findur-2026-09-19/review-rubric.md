# PRD High-Severity Recheck — Findur

> **Subsequent product override (2026-09-19):** The product owner judged the external-hosting approval gate disproportionate for an early demonstration. The PRD now permits hosting and sharing with clearly labeled draft content and retains qualified review as a public multi-user launch gate. This override supersedes the hosting-blocker resolution below; the compatibility recheck remains current.

## Overall verdict

Both former high-severity findings are resolved. No high-severity issue remains within this focused recheck.

## Public-hosting approval — SUPERSEDED BY PRODUCT OVERRIDE

**Original high finding:** The “publicly hosted demonstration” lacked an accountable approval gate for Terms, Privacy, and trust-and-safety content (§0, §11 Q8, FR-26).

**Resolution at the time of recheck:** §11 Q8 labeled this a **Milestone blocker**, made the product owner accountable, required qualified privacy/legal review plus product-safety review, rejected draft labels as sufficient, and placed the gate “before any external URL is enabled or shared.” The subsequent product override above replaces that resolution.

## Compatibility acceptance — PASS

**Original high finding:** Compatibility could pass without being correct because SM-2 required only some change in eligibility, order, or comparison, while similar/diversified/complementary had no expected direction (§4.4 FR-5, §4.5 FR-8 and FR-10, §7.1 SM-2, §11 Q1).

**Current resolution:** FR-5 now supplies seeded directional acceptance cases: similar favors closer composition, diversified favors broader asset-class and issuer distribution, and complementary favors lower overlap with balancing exposures. A-5 requires product confirmation before matching-story acceptance, while §11 Q1 assigns the exact signals and weights to product + architecture before matching implementation approval and requires the mechanism to satisfy those directional cases. The implementation details remain open without leaving acceptance direction undefined.

**Residual note:** SM-2 still uses broad change-only wording, but FR-5 now provides the missing correctness oracle and approval gate; this is not a remaining high-severity defect.
