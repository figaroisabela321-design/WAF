# Third-party: SamWaf

## Identity

| Field | Value |
|-------|-------|
| Project | SamWaf |
| Upstream | https://github.com/samwafgo/SamWaf |
| Pin SHA | `d975b12ec0a4757ca0e9698accd373dfee5f7c71` |
| Pin describe | `v1.3.25-beta.6-2-gd975b12` |
| License | Apache License 2.0 |
| Local freeze path | `/workspace/upstream/SamWaf` (not vendored into this repo) |

## Components of interest

| Component | Version / note |
|-----------|----------------|
| Coraza | `github.com/corazawaf/coraza/v3 v3.3.3` |
| OWASP CRS (bundled in SamWaf exedata) | `4.9.0-dev` |
| Go (upstream) | 1.25 / toolchain 1.25.9 |

## Usage in gov-waf Phase 2A

**Evaluation / analysis only.** No SamWaf source was copied into `gov-waf`. No runtime binary was shipped. `SAMWAF_RUNTIME_POC = BLOCKED`.

If a future phase reuses SamWaf code under Apache-2.0:

1. Preserve copyright and license headers on reused files.
2. Include this notice and the Apache-2.0 LICENSE text in distribution docs.
3. Document modifications clearly.

## Notice

This product evaluation may include references to SamWaf (Apache-2.0). SamWaf is Copyright its respective contributors. See upstream `LICENSE` and `ThirdLicense`.
