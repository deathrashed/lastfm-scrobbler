# First-run setup wizard

An interactive launch with no usable credential/configuration state opens the
wizard automatically. Run `scrobbler setup` to open it explicitly later.

The pages are:

1. System detection
2. Optional Nerd Font selection
3. Last.fm account
4. Credential storage
5. Recommended scrobbling preferences
6. Interface preferences
7. Review
8. Apply / Connection Test
9. Setup Complete

Review is the write boundary. Before Apply, credentials, configuration files,
terminal configuration, and font installation are untouched. Esc, Ctrl+C, or
skipping Welcome leaves no partial setup state. Existing working configuration
is not overwritten merely because the wizard is opened again.

The wizard detects macOS, Linux, and Windows context without requiring a
package manager. Credential choices reflect implemented backends: macOS
Keychain where available, otherwise the existing owner-only credentials file
or environment variables. Nerd Font installation is optional and uses an
official release archive at user scope. Terminal default configuration is only
automatic for explicitly supported adapters; unsupported terminals are marked
manual and do not make setup fail.

After Apply, the wizard tests the Last.fm API/authentication path and shows a
green or clearly failed step. Setup Complete continues to the normal
dashboard. Settings remains the place to change preferences later.
