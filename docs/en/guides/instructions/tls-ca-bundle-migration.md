# Instruction: TLS migration from `insecure` to `ca_bundle`

Language: [Русский](../../../ru/guides/instructions/tls-ca-bundle-migration.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Instructions](index.md)
      - [TLS: insecure → ca_bundle](tls-ca-bundle-migration.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

How to cleanly switch from `insecure: true` to a corporate `ca_bundle`
without downtime and without rewriting CI.

## When you need it

- Corporate TestRail behind a MITM proxy with a self-signed root CA.
- SRE requirement: drop `--insecure` / `insecure: true` in prod.
- You want to silence the `WARNING: TLS verification disabled` log
  banner without sacrificing security.

## Case A: Single corporate CA

**Source:**

```yaml
# ~/.gotr/config/default.yaml (v3.2-style)
insecure: true
```

**Steps:**

```bash
# 1. Get the CA PEM (from your SRE/PKI team)
sudo cp corporate-ca.pem /etc/ssl/corp-ca.pem
sudo chmod 0644 /etc/ssl/corp-ca.pem

# 2. Back up the config
cp ~/.gotr/config/default.yaml ~/.gotr/config/default.yaml.bak
```

**New config:**

```yaml
# ~/.gotr/config/default.yaml (v3.3-style)
tls:
  insecure: false
  ca_bundle: "/etc/ssl/corp-ca.pem"
```

**Validation:**

```bash
gotr self-test            # basic TLS handshake check
gotr get projects         # full-blown request
```

The `tls_insecure` banner **must not appear** on stderr.

## Case B: Keep insecure as a fallback on one machine

```yaml
tls:
  ca_bundle: "/etc/ssl/corp-ca.pem"
```

```bash
# On a machine where the CA is not rolled out yet — one-shot
gotr get projects --insecure
```

The top-level `insecure` and `--insecure` are OR-merged with
`tls.insecure` — existing scripts keep working.

## Case C: Suppress the banner when insecure is required in CI

```yaml
tls:
  insecure: true   # temporary measure — documented
ui:
  suppress_warnings: [tls_insecure]
```

```bash
# Show the banner locally, ignoring suppress
gotr get projects --show-warnings
```

## Case D: Multiple CAs

A PEM bundle is a concatenation of PEM certificates. Just stack them
into a single file:

```bash
cat corp-root.pem corp-intermediate.pem > /etc/ssl/corp-ca.pem
```

`x509.CertPool.AppendCertsFromPEM` reads every block.

## Troubleshooting

| Symptom | Cause | Fix |
| --- | --- | --- |
| `x509: certificate signed by unknown authority` | ca_bundle does not include the root CA | Add the root cert to the PEM |
| `open /etc/ssl/corp-ca.pem: permission denied` | chmod/ownership | `chmod 0644`, check the group |
| `failed to parse CA PEM` | empty or invalid file | `openssl x509 -in <file> -text -noout` |
| TLS banner still appears | `insecure=true` is still active somewhere | `gotr config view`, check the `TESTRAIL_INSECURE` env / legacy `insecure:` |

## Success criteria

- `gotr self-test` and `gotr get projects` exit 0 without a TLS warning.
- `go tool dist` / `openssl s_client` to the TestRail host succeed
  through the same PEM (sanity check).
- The config has no top-level `insecure: true`.

## See also

- [Configuration → TLS and ca_bundle](../configuration.md).
- [Migration guide v3.3 → TLS](../migration-guide-v3.3.md).
- [Architecture: UX polish v3.3.0 → ca_bundle](../../architecture/ux-polish-v3.3.0.md).

---

← [Instructions](index.md) · [Documentation](../../index.md)
