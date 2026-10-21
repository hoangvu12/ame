param([string]$Reference = '')
$ErrorActionPreference = 'Stop'
Push-Location $PSScriptRoot
try {
    cargo build --release --locked
    if ($LASTEXITCODE) { throw 'LTK overlay build failed' }
    $stage = Join-Path $PSScriptRoot 'build/bundle'
    New-Item -ItemType Directory -Force $stage | Out-Null
    $revision = '08468a32392f43dc604ebf7a8fcaf05c2bbeacb9'
    $expected = @{
        'ltk_patcher_host.exe' = '23fa1aaeda1a0c743da44227f179084e1fb81c7cd262144a3c8604ffc49d3bc1'
        'ltk_patcher_dll.dll' = '6d419057e6667994ba752ad0fb089b363db98267618644d7f7b6632441a21d74'
    }
    foreach ($name in $expected.Keys) {
        $target = Join-Path $stage $name
        if ($Reference) { Copy-Item -LiteralPath (Join-Path $Reference "src-tauri/resources/$name") -Destination $target -Force }
        else { Invoke-WebRequest -Uri "https://raw.githubusercontent.com/LeagueToolkit/ltk-manager/$revision/src-tauri/resources/$name" -OutFile $target }
        if ((Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected[$name]) { throw "Unexpected LTK binary: $name" }
    }
    if ((Get-AuthenticodeSignature (Join-Path $stage 'ltk_patcher_dll.dll')).Status -ne 'Valid') { throw 'LTK DLL signature invalid' }
    Copy-Item -LiteralPath target/release/ame-ltk-overlay.exe -Destination $stage -Force
    if ($Reference) { Copy-Item -LiteralPath (Join-Path $Reference 'LICENSE') -Destination (Join-Path $stage 'LTK-LICENSE') -Force }
    else { Invoke-WebRequest -Uri "https://raw.githubusercontent.com/LeagueToolkit/ltk-manager/$revision/LICENSE" -OutFile (Join-Path $stage 'LTK-LICENSE') }
    "LTK manager source: https://github.com/LeagueToolkit/ltk-manager/tree/$revision`nOverlay helper source: tools/ltk-runtime (Cargo.lock pins dependencies). Signed runtime binaries are unmodified." | Set-Content (Join-Path $stage 'NOTICE.txt')
    New-Item -ItemType Directory -Force (Join-Path $stage 'source/src') | Out-Null
    Copy-Item Cargo.toml,Cargo.lock,build.ps1 -Destination (Join-Path $stage 'source') -Force
    Copy-Item src/main.rs -Destination (Join-Path $stage 'source/src') -Force
    Compress-Archive -Path "$stage/*" -DestinationPath ../../internal/ltk/assets/runtime.zip -Force
} finally { Pop-Location }
