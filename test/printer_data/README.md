# ⚠️ Test Environment Only — Not for Production Use

This directory (`test/printer_data/`) contains configuration files for the
**virtual test environment** only. It is not a starter config or reference
configuration for a real Klipper/Moonraker deployment.

**Security notice:** The `moonraker.conf` in this directory intentionally
disables authentication and permits unrestricted access:

| Setting | Value | Risk |
|---------|-------|------|
| `force_logins` | `False` | No API authentication |
| `trusted_clients` | `0.0.0.0/0` | Any IP can connect |
| `cors_domains` | `*` | Any origin can make requests |

**Never** copy these files to a production Klipper host. Doing so will expose
your entire Klipper/Moonraker stack without any authentication or access
controls.

For a production Moonraker configuration, refer to the official documentation:
https://moonraker.readthedocs.io/en/latest/configuration/

---

For details on running the test environment, see the top-level `test/README.md`.
