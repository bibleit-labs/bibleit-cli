param(
    [string]$Version = '',
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA 'Bibleit\bin'),
    [string]$AssetsDir = ''
)
# Windows PowerShell 5.1 and PowerShell 7. No elevation or profile edits required.
$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'This installer requires Windows' }
$repo = 'https://github.com/bibleit-labs/bibleit-cli'
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
if (-not $Version) {
    if ($AssetsDir) { throw 'Specify -Version for local assets' }
    $Version = (Invoke-RestMethod 'https://api.github.com/repos/bibleit-labs/bibleit-cli/releases/latest').tag_name
}
if ($Version -cnotmatch '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[A-Za-z0-9]+([.-][A-Za-z0-9]+)*)?$') { throw 'Invalid version' }
$cpu = $env:PROCESSOR_ARCHITEW6432
if (-not $cpu) { $cpu = $env:PROCESSOR_ARCHITECTURE }
switch ($cpu) { 'AMD64' { $arch = 'amd64' }; 'ARM64' { $arch = 'arm64' }; default { throw "Unsupported CPU: $cpu" } }
$name = "bibleit_$($Version.Substring(1))_windows_$arch.zip"
$temporary = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString())
$staged = $null
New-Item -ItemType Directory -Path $temporary | Out-Null
try {
    foreach ($asset in @($name, 'checksums.txt')) {
        $path = Join-Path $temporary $asset
        if ($AssetsDir) { Copy-Item -LiteralPath (Join-Path $AssetsDir $asset) -Destination $path }
        else { Invoke-WebRequest -UseBasicParsing -Uri "$repo/releases/download/$Version/$asset" -OutFile $path }
    }
    $entries = @(Get-Content -LiteralPath (Join-Path $temporary 'checksums.txt') | Where-Object { $_ -cmatch ('^[0-9a-f]{64}  ' + [regex]::Escape($name) + '$') })
    if ($entries.Count -ne 1) { throw 'Missing or duplicate SHA-256 checksum' }
    $expected = $entries[0].Substring(0, 64)
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $temporary $name)).Hash.ToLowerInvariant()
    if ($actual -cne $expected) { throw 'Checksum mismatch; existing installation unchanged' }
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $zip = [IO.Compression.ZipFile]::OpenRead((Join-Path $temporary $name))
    try {
        $binaries = @($zip.Entries | Where-Object { $_.FullName -ceq 'bibleit.exe' })
        if ($binaries.Count -ne 1) { throw 'Archive contains no unique binary' }
        $binary = Join-Path $temporary 'bibleit.exe'
        [IO.Compression.ZipFileExtensions]::ExtractToFile($binaries[0], $binary)
    } finally { $zip.Dispose() }
    $reported = & $binary version
    if ($LASTEXITCODE -ne 0 -or $reported -cne "bibleit CLI version $($Version.Substring(1))") { throw 'Binary version mismatch' }
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $staged = Join-Path $InstallDir ('.bibleit-' + [Guid]::NewGuid().ToString() + '.exe')
    Copy-Item -LiteralPath $binary -Destination $staged
    $target = Join-Path $InstallDir 'bibleit.exe'
    if (Test-Path -LiteralPath $target) { [IO.File]::Replace($staged, $target, $null) }
    else { [IO.File]::Move($staged, $target) }
    Write-Output "Installed Bibleit $Version at $target. Add $InstallDir to your user PATH if needed."
} finally {
    Remove-Item -LiteralPath $temporary -Recurse -Force
    if ($staged -and (Test-Path -LiteralPath $staged)) { Remove-Item -LiteralPath $staged -Force }
}
