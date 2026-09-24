param([string]$Reference = '')
$ErrorActionPreference = 'Stop'
Push-Location $PSScriptRoot
try {
    cargo build --release --locked
    if ($LASTEXITCODE) { throw 'LTK overlay build failed' }
    $stage = Join-Path $PSScriptRoot 'build/bundle'
    New-Item -ItemType Directory -Force $stage | Out-Null
    $revision = 'ff7f2ebc0e2abccd280d8634bb015697400e0121'
    $expected = @{
        'ltk_patcher_host.exe' = 'a7c4047ce7548c7ae820bc440735f15b9d1a495acf061dbb5a5a2893a0ed8d7c'
        'ltk_patcher_dll.dll' = '07a43bf36a389eb00f6276e333bd7f2b95218f25a58e1e128ff4d2e4ab2dc99b'
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
