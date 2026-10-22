$ErrorActionPreference = 'Stop'

# Removes the shim(s) created for the binaries extracted from the release
# archive. Chocolatey removes the extracted files from the package directory
# automatically on uninstall.
Uninstall-ChocolateyZipPackage -PackageName 'unirate-cli' -ZipFileName 'unirate_0.1.0_windows_amd64.zip'
