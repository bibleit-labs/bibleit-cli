param([Parameter(Mandatory = $true)][string]$AssetsDir)
# Parser and failure-path checks also run in a Linux PowerShell container.
# This does not substitute for native Windows installation/replacement checks.
$ErrorActionPreference = 'Stop'
$installer = Join-Path (Split-Path $PSScriptRoot -Parent) 'install.ps1'
$tokens = $null
$errors = $null
[System.Management.Automation.Language.Parser]::ParseFile($installer, [ref]$tokens, [ref]$errors) | Out-Null
if ($errors.Count) { throw ($errors | Out-String) }
$temporary = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString())
$originalOS = $env:OS
$originalCPU = $env:PROCESSOR_ARCHITECTURE
$originalCPU64 = $env:PROCESSOR_ARCHITEW6432
try {
    $env:OS = 'Windows_NT'
    $env:PROCESSOR_ARCHITECTURE = 'AMD64'
    $env:PROCESSOR_ARCHITEW6432 = ''
    $assets = Join-Path $temporary 'assets'
    $destination = Join-Path $temporary 'installed'
    New-Item -ItemType Directory -Path $assets, $destination -Force | Out-Null
    Copy-Item -Path (Join-Path $AssetsDir '*') -Destination $assets
    $version = (Get-Content -LiteralPath (Join-Path $assets 'manifest.json') -Raw | ConvertFrom-Json).version
    $binary = Join-Path $destination 'bibleit.exe'
    Set-Content -LiteralPath $binary -Value 'existing installation'
    $previous = (Get-FileHash -LiteralPath $binary).Hash
    function Expect-Failure([string]$Requested, [string]$Message) {
        $failure = $null
        try { & $installer -Version $Requested -AssetsDir $assets -InstallDir $destination }
        catch { $failure = $_.Exception.Message }
        if (-not $failure -or $failure -notmatch $Message) { throw "Expected $Message; received $failure" }
        if ((Get-FileHash -LiteralPath $binary).Hash -ne $previous) { throw 'Existing installation modified' }
    }
    Expect-Failure '../invalid' 'Invalid version'
    $checksumPath = Join-Path $assets 'checksums.txt'
    $checksums = Get-Content -LiteralPath $checksumPath -Raw
    Set-Content -LiteralPath $checksumPath -Value ''
    Expect-Failure $version 'Missing or duplicate'
    Set-Content -LiteralPath $checksumPath -Value ($checksums + $checksums)
    Expect-Failure $version 'Missing or duplicate'
    Set-Content -LiteralPath $checksumPath -Value $checksums
    Get-ChildItem -LiteralPath $assets -Filter '*.zip' | ForEach-Object { [IO.File]::AppendAllText($_.FullName, 'corrupted') }
    Expect-Failure $version 'Checksum mismatch'
    Write-Output 'PASS: PowerShell syntax and invalid-version/missing-checksum/duplicate-checksum/corrupt-archive rejection; existing installation preserved.'
} finally {
    $env:OS = $originalOS
    $env:PROCESSOR_ARCHITECTURE = $originalCPU
    $env:PROCESSOR_ARCHITEW6432 = $originalCPU64
    if (Test-Path -LiteralPath $temporary) { Remove-Item -LiteralPath $temporary -Recurse -Force }
}
