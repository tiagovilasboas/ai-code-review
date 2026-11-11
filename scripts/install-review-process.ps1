# Install the multi-agent review process into a target repository (Windows PowerShell 5.1+ / pwsh).
# Usage: install-review-process.ps1 -Harness <claude-code|kiro|codex|cursor|copilot|gemini|all> -Target <dir> [-DryRun] [-Force] [-WithSecurityKit]
# Never overwrites an existing file unless -Force. -DryRun prints the plan only.
param(
  [Parameter(Mandatory = $true)][string]$Harness,
  [Parameter(Mandatory = $true)][string]$Target,
  [switch]$DryRun,
  [switch]$Force,
  [switch]$WithSecurityKit
)
$ErrorActionPreference = 'Stop'
$Src = Split-Path -Parent $PSScriptRoot
$All = @('claude-code', 'kiro', 'codex', 'cursor', 'copilot', 'gemini')
if ($Harness -eq 'all') { $List = $All }
elseif ($All -contains $Harness) { $List = @($Harness) }
else { throw "unknown harness: $Harness" }

if (-not $DryRun) { New-Item -ItemType Directory -Force -Path $Target | Out-Null }
if (Test-Path $Target) { $Target = (Resolve-Path $Target).Path }
$Begin = '<!-- review-process:begin -->'
$End = '<!-- review-process:end -->'
$script:Conflicts = 0

function Same($a, $b) { (Get-FileHash $a).Hash -eq (Get-FileHash $b).Hash }

function Copy-One($src, $rel) {
  $dst = Join-Path $Target ($rel -replace '/', [IO.Path]::DirectorySeparatorChar)
  if (Test-Path $dst) {
    if (Same $src $dst) { Write-Output "skip      $rel (identical)"; return }
    if (-not $Force) { Write-Output "conflict  $rel (exists; use -Force to overwrite)"; $script:Conflicts++; return }
    Write-Output "overwrite $rel"
  } else { Write-Output "create    $rel" }
  if ($DryRun) { return }
  New-Item -ItemType Directory -Force -Path (Split-Path -Parent $dst) | Out-Null
  Copy-Item -LiteralPath $src -Destination $dst -Force
}

function Copy-Tree($base, $out) {
  if (-not (Test-Path $base)) { Write-Output "warn      missing source dir $base"; return }
  $root = (Resolve-Path $base).Path
  Get-ChildItem -Path $root -Recurse -File -Force | Sort-Object FullName | ForEach-Object {
    $rel = $_.FullName.Substring($root.Length).TrimStart('\', '/') -replace '\\', '/'
    Copy-One $_.FullName "$out/$rel"
  }
}

function Snippet-One($src, $rel) {
  $dst = Join-Path $Target ($rel -replace '/', [IO.Path]::DirectorySeparatorChar)
  $text = Get-Content -Raw -LiteralPath $src
  if (-not (Test-Path $dst)) {
    Write-Output "create    $rel (snippet)"
    if (-not $DryRun) {
      New-Item -ItemType Directory -Force -Path (Split-Path -Parent $dst) | Out-Null
      Set-Content -LiteralPath $dst -Value $text -NoNewline
    }
    return
  }
  $cur = Get-Content -Raw -LiteralPath $dst
  if ($cur.Contains($Begin)) {
    if (-not $Force) { Write-Output "skip      $rel (block already present)"; return }
    Write-Output "replace   $rel (snippet block)"
    if (-not $DryRun) {
      $pattern = '(?s)' + [regex]::Escape($Begin) + '.*?' + [regex]::Escape($End)
      $new = [regex]::Replace($cur, $pattern, { param($m) $text.Trim() })
      Set-Content -LiteralPath $dst -Value $new -NoNewline
    }
    return
  }
  Write-Output "append    $rel (snippet block)"
  if (-not $DryRun) { Add-Content -LiteralPath $dst -Value ("`n" + $text) -NoNewline }
}

if ($DryRun) { Write-Output "dry-run: nothing will be written (target: $Target)" }
Copy-Tree (Join-Path $Src 'reviewers') 'reviewers'
Copy-One (Join-Path $Src '.agents/skills/pr-review-process/SKILL.md') '.agents/skills/pr-review-process/SKILL.md'
$rv = Join-Path $Src 'templates/REVIEW.md'
if (Test-Path $rv) { Copy-One $rv 'REVIEW.md' } else { Write-Output 'warn      templates/REVIEW.md not found; REVIEW.md not installed' }
$rt = Join-Path $Src 'templates/review-tiers.json'
if (Test-Path $rt) { Copy-One $rt '.github/review-tiers.json' } else { Write-Output 'warn      templates/review-tiers.json not found; tiers config not installed' }
if ($WithSecurityKit) {
  Copy-Tree (Join-Path $Src 'guardrails') 'guardrails'
  Copy-Tree (Join-Path $Src 'skills') 'skills'
  Copy-One (Join-Path $Src '.agents/skills/appsec-pr-review/SKILL.md') '.agents/skills/appsec-pr-review/SKILL.md'
}
foreach ($h in $List) {
  $hdir = Join-Path $Src "harness/$h"
  if (-not (Test-Path $hdir)) { Write-Output "warn      missing harness dir harness/$h"; continue }
  $root = (Resolve-Path $hdir).Path
  Get-ChildItem -Path $root -Recurse -File -Force | Sort-Object FullName | ForEach-Object {
    $rel = $_.FullName.Substring($root.Length).TrimStart('\', '/') -replace '\\', '/'
    if ($rel -like '.*/*') { Copy-One $_.FullName $rel }
  }
  switch ($h) {
    'codex' { Snippet-One (Join-Path $hdir 'AGENTS.snippet.md') 'AGENTS.md' }
    'copilot' { Snippet-One (Join-Path $hdir 'copilot-instructions.snippet.md') '.github/copilot-instructions.md' }
    'gemini' { Snippet-One (Join-Path $hdir 'GEMINI.snippet.md') 'GEMINI.md' }
  }
}
if ($script:Conflicts -gt 0) {
  Write-Output "$($script:Conflicts) file(s) kept as they were. Re-run with -Force to overwrite."
  exit 3
}
if (-not $DryRun) { Write-Output 'done. Next: review REVIEW.md, edit .github/review-tiers.json, then run the review command for your harness.' }
