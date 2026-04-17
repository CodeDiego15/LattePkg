# Example manifests

These files illustrate the Armada manifest format. They use real URLs but the
`checksum:` fields are placeholders (`REPLACE_WITH_REAL_SHA256`) so that you
can review the structure without accidentally installing tools that have not
been checksummed against the real releases.

To use these for real installs, replace the placeholder checksums with the
actual SHA-256 digests published by the upstream project.

The `nmap.yaml` file is kept in the legacy `platforms:` format to demonstrate
that `armada simulate` still understands older manifests.

Modern manifests (`ffuf.yaml`, `httpx.yaml`, `nuclei.yaml`, `subfinder.yaml`,
`gobuster.yaml`, `amass.yaml`) use the new `assets:` map which is what the
real installer consumes.
