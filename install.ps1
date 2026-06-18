#Requires -Version 5.0
$ErrorActionPreference = "Stop"

$Repo    = "alvaroeng98/harness-init"
$Binary  = "harness-init"
$InstDir = "$env:LOCALAPPDATA\Programs\harness-init"

$Arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Error "Solo se soporta 64-bit."; exit 1
}

if (-not $env:VERSION) {
    $rel = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $rel.tag_name
} else {
    $Version = $env:VERSION
}

$Url  = "https://github.com/$Repo/releases/download/$Version/$Binary-windows-$Arch.exe"
$Dest = "$InstDir\$Binary.exe"

Write-Host "Instalando $Binary $Version (windows/$Arch)..."
New-Item -ItemType Directory -Force -Path $InstDir | Out-Null
Invoke-WebRequest -Uri $Url -OutFile $Dest

$path = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($path -notlike "*$InstDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$path;$InstDir", "User")
    Write-Host "Añadido $InstDir al PATH de usuario."
}

Write-Host "Instalado en $Dest"
Write-Host "Reinicia la terminal y ejecuta: harness-init --help"
