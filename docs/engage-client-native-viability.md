# Engage client-native — viability (Go / No-Go)

This document captures the **product decision** after the user-friendly install + red-vs-blue lab track.

**Status:** TBD — complete after field testing and fix slices.

## Criteria (from plan)

- **Go:** install + preflight succeed on at least two package-manager families; red harness completes without killing the victim process; critical issues have documented mitigations.
- **No-Go:** core tools cannot be installed on non-Kali without excessive manual steps; victim crashes on baseline harness; package matrix cost outweighs benefit.

## Evidence checklist

- [ ] Distro matrix tried (record in [engage-red-blue-bugs.md](engage-red-blue-bugs.md))
- [ ] `make engage-install-plan` reviewed
- [ ] `make engage-install-host-tools` (or minimal profile) outcome
- [ ] `./scripts/engage/preflight-client-tools.sh --json` result archived
- [ ] `make test-engage-red-blue` against local victim

## Decision

| Verdict | Owner | Date | Notes |
|---------|-------|------|-------|
| Go / No-Go |       |      |       |
