# Updates

Official builds use the project's GitHub Releases source by default:

```text
github.com/deathrashed/lastfm-scrobbler
```

No `SCROBBLER_UPDATE_URL` value is required for normal update checks.

Use either surface:

```text
scrobbler check-update
Settings → Tools → Check for Updates
```

The checker reads the latest release metadata and reports whether a newer
version exists; it does not replace the running binary. A custom endpoint can
be supplied through `SCROBBLER_UPDATE_URL` or the Settings → Tools update
source field. Custom responses may use GitHub-style `tag_name`, `html_url`,
and `body` fields or the simpler `version`, `url`, and `notes` fields.

An empty custom value restores the built-in official repository source.
Network errors, malformed URLs, non-success HTTP responses, and responses
without a version are reported as update-check failures without exposing
credentials.
