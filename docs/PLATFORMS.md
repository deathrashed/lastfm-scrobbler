# Platform support

The application targets macOS, Linux, and Windows. The shared TUI and CLI do
not require a package manager.

| Capability | macOS | Linux | Windows |
| --- | --- | --- | --- |
| TUI / CLI | supported | supported | supported |
| First-run wizard | supported | supported | supported |
| User-scope Nerd Font install | supported | supported | supported |
| Credential storage | Keychain, file, environment | file, environment | file, environment |
| Native/desktop picker | native picker | desktop tool when available; manual fallback | Windows Forms picker via PowerShell |
| Notifications | supported where macOS notification API is available | unavailable | unavailable |
| Terminal auto-configuration | supported adapters only | manual unless an adapter is added | manual unless an adapter is added |
| Completion installation | supported | supported | supported |
| GitHub Releases update checks | supported | supported | supported |

Linux picker availability depends on the desktop tools installed by the user;
the application detects them rather than assuming a desktop environment.
Credential storage does not claim a Linux Secret Service or Windows Credential
Manager adapter. Terminal font configuration does not edit arbitrary terminal
files. Unsupported integrations are reported as manual or unavailable while
the core File, CLI, and TUI workflows remain usable.
