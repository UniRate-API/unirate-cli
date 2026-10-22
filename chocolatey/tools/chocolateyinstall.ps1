$ErrorActionPreference = 'Stop'
$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition

# Downloads the official, checksum-pinned UniRate CLI release archive from
# GitHub and extracts the binary into the package tools directory, where
# Chocolatey auto-shims `unirate.exe` onto the PATH. The amd64 build runs
# natively on x64 Windows and under x64 emulation on Windows on ARM.
$packageArgs = @{
  packageName    = 'unirate-cli'
  unzipLocation  = $toolsDir
  url64bit       = 'https://github.com/UniRate-API/unirate-cli/releases/download/v0.1.0/unirate_0.1.0_windows_amd64.zip'
  checksum64     = '667f111dcaf53fea600e677dfeb8871a32eb019137ef2a407adc4e0cefb1957f'
  checksumType64 = 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
