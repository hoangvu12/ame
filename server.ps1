$ErrorActionPreference = 'Stop'

$AME_DIR = "$env:LOCALAPPDATA\ame"
$TOOLS_DIR = "$AME_DIR\tools"
$SKINS_DIR = "$AME_DIR\skins"
$MODS_DIR = "$AME_DIR\mods"
$OVERLAY_DIR = "$AME_DIR\overlay"
$SKIN_BASE_URL = "https://raw.githubusercontent.com/Alban1911/LeagueSkins/main/skins"
$PORT = 18765

# --- Game directory detection ---
function Find-GameDir {
    # 1. Common paths on all drives
    foreach ($drive in [System.IO.DriveInfo]::GetDrives() | Where-Object { $_.DriveType -eq 'Fixed' }) {
        $path = Join-Path $drive.RootDirectory.FullName "Riot Games\League of Legends\Game"
        if (Test-Path (Join-Path $path "League of Legends.exe")) { return $path }
    }

    # 2. RiotClientInstalls.json
    $rcPath = "C:\ProgramData\Riot Games\RiotClientInstalls.json"
    if (Test-Path $rcPath) {
        try {
            $rc = Get-Content $rcPath -Raw | ConvertFrom-Json
            $paths = @()
            if ($rc.associated_client) {
                $rc.associated_client.PSObject.Properties | ForEach-Object { $paths += $_.Name }
            }
            if ($rc.rc_default) { $paths += $rc.rc_default }
            if ($rc.rc_live) { $paths += $rc.rc_live }
            foreach ($p in $paths) {
                if ($p -match '(?i)league') {
                    $candidate = Split-Path $p -Parent
                    $gameDir = Join-Path $candidate "Game"
                    if (Test-Path (Join-Path $gameDir "League of Legends.exe")) { return $gameDir }
                }
            }
        } catch {}
    }

    # 3. Running LeagueClientUx.exe
    try {
        $proc = Get-CimInstance Win32_Process -Filter "Name='LeagueClientUx.exe'" -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($proc) {
            $gameDir = Join-Path (Split-Path $proc.ExecutablePath -Parent) "Game"
            if (Test-Path (Join-Path $gameDir "League of Legends.exe")) { return $gameDir }
        }
    } catch {}

    # 4. Registry
    try {
        $regPath = "HKLM:\SOFTWARE\WOW6432Node\Riot Games, Inc\League of Legends"
        $loc = (Get-ItemProperty -Path $regPath -Name "Location" -ErrorAction SilentlyContinue).Location
        if ($loc) {
            $gameDir = Join-Path $loc "Game"
            if (Test-Path (Join-Path $gameDir "League of Legends.exe")) { return $gameDir }
        }
    } catch {}

    return $null
}

# --- WebSocket helpers ---
function Send-WsMessage($ws, $obj) {
    if ($ws.State -ne [System.Net.WebSockets.WebSocketState]::Open) { return }
    $json = $obj | ConvertTo-Json -Compress
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($json)
    $segment = [System.ArraySegment[byte]]::new($bytes)
    $ws.SendAsync($segment, [System.Net.WebSockets.WebSocketMessageType]::Text, $true, [System.Threading.CancellationToken]::None).GetAwaiter().GetResult() | Out-Null
}

function Send-Status($ws, $status, $message) {
    Send-WsMessage $ws @{ type = "status"; status = $status; message = $message }
    Write-Host "[ame] $status`: $message"
}


# --- Skin apply ---

function Invoke-Apply($ws, $championId, $skinId) {
    # Called when user clicks "Apply Skin" — do everything now
    $gameDir = Find-GameDir
    if (-not $gameDir) {
        Send-Status $ws "error" "League of Legends Game directory not found"
        return
    }
    Write-Host "[ame] Game dir: $gameDir"

    $modTools = Join-Path $TOOLS_DIR "mod-tools.exe"
    if (-not (Test-Path $modTools)) {
        Send-Status $ws "error" "mod-tools.exe not found. Run install.bat first."
        return
    }

    # Download skin file if not cached (.fantome or .zip)
    $skinDir = Join-Path $SKINS_DIR "$championId\$skinId"
    $zipPath = Join-Path $skinDir "$skinId.fantome"

    if (-not (Test-Path $zipPath)) {
        # Also check for .zip from previous downloads
        $altPath = Join-Path $skinDir "$skinId.zip"
        if (Test-Path $altPath) {
            $zipPath = $altPath
        }
    }

    if (-not (Test-Path $zipPath)) {
        Send-Status $ws "downloading" "Downloading skin..."
        New-Item -ItemType Directory -Path $skinDir -Force | Out-Null
        $downloaded = $false
        foreach ($ext in @("fantome", "zip")) {
            $url = "$SKIN_BASE_URL/$championId/$skinId/$skinId.$ext"
            $tryPath = Join-Path $skinDir "$skinId.$ext"
            try {
                Invoke-WebRequest -Uri $url -OutFile $tryPath -UseBasicParsing
                $zipPath = $tryPath
                $downloaded = $true
                Write-Host "[ame] Downloaded $skinId.$ext"
                break
            } catch {
                Remove-Item $tryPath -Force -ErrorAction SilentlyContinue
            }
        }
        if (-not $downloaded) {
            Send-Status $ws "error" "Skin not available for download"
            return
        }
    } else {
        Write-Host "[ame] Skin $skinId already cached"
    }

    # Kill any previous runoverlay
    Get-Process -Name "mod-tools" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Milliseconds 300

    # Clean and extract to mods dir (into a named subfolder so mod-tools finds it)
    if (Test-Path $MODS_DIR) { Remove-Item $MODS_DIR -Recurse -Force }
    $modSubDir = Join-Path $MODS_DIR "skin_$skinId"
    New-Item -ItemType Directory -Path $modSubDir -Force | Out-Null
    Expand-Archive -Path $zipPath -DestinationPath $modSubDir -Force

    # Clean overlay dir
    if (Test-Path $OVERLAY_DIR) { Remove-Item $OVERLAY_DIR -Recurse -Force }
    New-Item -ItemType Directory -Path $OVERLAY_DIR -Force | Out-Null

    # Run mkoverlay
    Send-Status $ws "injecting" "Applying skin..."
    $modName = "skin_$skinId"
    $mkArgs = "mkoverlay `"$MODS_DIR`" `"$OVERLAY_DIR`" `"--game:$gameDir`" --mods:$modName --noTFT --ignoreConflict"
    Write-Host "[ame] Running: mod-tools.exe $mkArgs"
    $mkProc = Start-Process -FilePath $modTools -ArgumentList $mkArgs -NoNewWindow -Wait -PassThru

    if ($mkProc.ExitCode -ne 0) {
        Send-Status $ws "error" "Failed to apply skin (exit code $($mkProc.ExitCode))"
        return
    }

    # Run runoverlay immediately
    $configPath = Join-Path $OVERLAY_DIR "cslol-config.json"
    $runArgs = "runoverlay `"$OVERLAY_DIR`" `"$configPath`" `"--game:$gameDir`" --opts:configless"
    Write-Host "[ame] Running: mod-tools.exe $runArgs"
    Start-Process -FilePath $modTools -ArgumentList $runArgs -NoNewWindow

    Send-Status $ws "ready" "Skin applied!"
}

function Invoke-Cleanup {
    Write-Host "[ame] Cleanup: stopping runoverlay"
    Get-Process -Name "mod-tools" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    if (Test-Path $OVERLAY_DIR) { Remove-Item $OVERLAY_DIR -Recurse -Force -ErrorAction SilentlyContinue }
}

# --- Main server loop ---
$listener = [System.Net.HttpListener]::new()
$listener.Prefixes.Add("http://localhost:$PORT/")
$listener.Start()
Write-Host "[ame] Listening on ws://localhost:$PORT"

try {
    while ($true) {
        $ctx = $listener.GetContext()

        if (-not $ctx.Request.IsWebSocketRequest) {
            $ctx.Response.StatusCode = 400
            $ctx.Response.Close()
            continue
        }

        try {
            $wsCtx = $ctx.AcceptWebSocketAsync([NullString]::Value).GetAwaiter().GetResult()
        } catch {
            Write-Host "[ame] WebSocket accept failed: $_"
            try { $ctx.Response.Close() } catch {}
            continue
        }
        $ws = $wsCtx.WebSocket
        Write-Host "[ame] Client connected"

        $buf = [byte[]]::new(4096)

        while ($ws.State -eq [System.Net.WebSockets.WebSocketState]::Open) {
            try {
                $segment = [System.ArraySegment[byte]]::new($buf)
                $result = $ws.ReceiveAsync($segment, [System.Threading.CancellationToken]::None).GetAwaiter().GetResult()

                if ($result.MessageType -eq [System.Net.WebSockets.WebSocketMessageType]::Close) {
                    $ws.CloseAsync([System.Net.WebSockets.WebSocketCloseStatus]::NormalClosure, "", [System.Threading.CancellationToken]::None).GetAwaiter().GetResult() | Out-Null
                    break
                }

                $json = [System.Text.Encoding]::UTF8.GetString($buf, 0, $result.Count)
                $msg = $json | ConvertFrom-Json

                switch ($msg.type) {
                    "apply" {
                        Write-Host "[ame] Apply requested: champion=$($msg.championId) skin=$($msg.skinId)"
                        Invoke-Apply $ws $msg.championId $msg.skinId
                    }
                    "cleanup" {
                        Invoke-Cleanup
                    }
                    default {
                        Write-Host "[ame] Unknown message type: $($msg.type)"
                    }
                }
            } catch {
                Write-Host "[ame] WebSocket error: $_"
                break
            }
        }
        Write-Host "[ame] Client disconnected"
    }
} finally {
    $listener.Stop()
    Invoke-Cleanup
    # Kill Pengu Loader when server exits
    Get-Process -Name "Pengu Loader" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    Write-Host "[ame] Server stopped"
}
