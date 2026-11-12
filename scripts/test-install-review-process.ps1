# Exercises scripts/install-review-process.ps1 the way test-install-review-process.sh
# exercises the shell installer: dry-run, install, idempotency, refusal to overwrite,
# security kit, and the same file list as the shell installer when sh is available.
# Usage: pwsh -NoProfile -File scripts/test-install-review-process.ps1
$ErrorActionPreference = 'Stop'
$Installer = Join-Path $PSScriptRoot 'install-review-process.ps1'
$Tmp = Join-Path ([IO.Path]::GetTempPath()) ('review-process-' + [guid]::NewGuid())
New-Item -ItemType Directory -Path $Tmp | Out-Null
$script:Fails = 0

function Check($name, $ok) {
  if ($ok) { Write-Output "ok   $name" } else { Write-Output "FAIL $name"; $script:Fails++ }
}
# The installer ends with `exit`, so each run is its own process.
function Install { pwsh -NoProfile -File $Installer @args 2>&1 | ForEach-Object { "$_" } }
function Files($dir) {
  $root = (Resolve-Path -LiteralPath $dir).Path
  Get-ChildItem -LiteralPath $root -Recurse -File -Force |
    ForEach-Object { $_.FullName.Substring($root.Length).TrimStart('\', '/') -replace '\\', '/' } | Sort-Object
}

try {
  # 1. dry-run writes nothing
  $dry = Join-Path $Tmp 'dry'
  $out = Install -Harness all -Target $dry -DryRun
  Check 'dry-run prints a plan' ($out -contains 'create    reviewers/bug-hunter.md')
  Check 'dry-run writes no files' (-not (Test-Path -LiteralPath $dry))

  # 2. real install, every harness
  $repo = Join-Path $Tmp 'repo'
  Install -Harness all -Target $repo | Out-Null
  Check 'install exits 0' ($LASTEXITCODE -eq 0)
  $expected = @(
    'reviewers/bug-hunter.md', 'reviewers/security.md', 'reviewers/guardrails/untrusted-diff.md',
    '.agents/skills/pr-review-process/SKILL.md', 'REVIEW.md', '.github/review-tiers.json',
    '.claude/agents/bug-hunter.md', '.claude/commands/review-pr.md',
    '.kiro/steering/review-process.md', '.kiro/steering/review-pr.md', '.kiro/hooks/pre-pr-self-check.json',
    '.codex/prompts/review-pr.md', 'AGENTS.md', '.cursor/rules/review-process.mdc',
    '.github/copilot-instructions.md', '.github/prompts/review-pr.prompt.md',
    'GEMINI.md', '.gemini/commands/review-pr.toml')
  foreach ($f in $expected) { Check "installed $f" (Test-Path -LiteralPath (Join-Path $repo $f) -PathType Leaf) }
  Check 'README of adapters not installed' (-not (Test-Path -LiteralPath (Join-Path $repo 'README.md')))
  Check 'security kit not installed without the flag' (-not (Test-Path -LiteralPath (Join-Path $repo 'security')))
  $bytes = [IO.File]::ReadAllBytes((Join-Path $repo 'AGENTS.md'))
  Check 'snippet file has no BOM' (-not ($bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF))

  # 3. second run changes nothing
  $out = Install -Harness all -Target $repo
  Check 'second run exits 0' ($LASTEXITCODE -eq 0)
  Check 'second run is stable' (-not ($out -match '^(create|overwrite|append|replace|conflict)'))

  # 4. a modified file is kept unless -Force
  $hunter = Join-Path $repo 'reviewers/bug-hunter.md'
  [IO.File]::WriteAllText($hunter, 'local edit')
  Install -Harness claude-code -Target $repo | Out-Null
  Check 'conflict exits 3' ($LASTEXITCODE -eq 3)
  Check 'modified file kept' ([IO.File]::ReadAllText($hunter) -eq 'local edit')
  Install -Harness claude-code -Target $repo -Force | Out-Null
  Check '-Force exits 0' ($LASTEXITCODE -eq 0)
  Check '-Force restores file' ([IO.File]::ReadAllText($hunter) -ne 'local edit')

  # 5. a begin marker without an end marker is never rewritten
  $agents = Join-Path $repo 'AGENTS.md'
  [IO.File]::WriteAllText($agents, "mine`n<!-- review-process:begin -->`nhalf a block`n")
  Install -Harness codex -Target $repo -Force | Out-Null
  Check 'begin without end under -Force exits 3' ($LASTEXITCODE -eq 3)
  Check 'begin without end leaves file unchanged' ([IO.File]::ReadAllText($agents) -eq "mine`n<!-- review-process:begin -->`nhalf a block`n")

  # 6. security kit and unknown harness
  $kit = Join-Path $Tmp 'kit'
  Install -Harness claude-code -Target $kit -WithSecurityKit | Out-Null
  Check 'security kit installed' (Test-Path -LiteralPath (Join-Path $kit 'security/skills/authz-idor.md'))
  Install -Harness nope -Target (Join-Path $Tmp 'nope') | Out-Null
  Check 'unknown harness fails' ($LASTEXITCODE -ne 0)

  # 7. same files as the shell installer
  if (Get-Command sh -ErrorAction SilentlyContinue) {
    $a = Join-Path $Tmp 'ps'; $b = Join-Path $Tmp 'sh'
    Install -Harness all -Target $a -WithSecurityKit | Out-Null
    sh (Join-Path $PSScriptRoot 'install-review-process.sh') --harness all --target ($b -replace '\\', '/') --with-security-kit | Out-Null
    $diff = Compare-Object (Files $a) (Files $b)
    if ($diff) { $diff | ForEach-Object { Write-Output "     $($_.SideIndicator) $($_.InputObject)" } }
    Check 'same file list as the shell installer' (-not $diff)
  } else { Write-Output 'skip same file list as the shell installer (no sh)' }
} finally {
  Remove-Item -LiteralPath $Tmp -Recurse -Force -ErrorAction SilentlyContinue
}
if ($script:Fails -gt 0) { Write-Output "$($script:Fails) check(s) failed"; exit 1 }
Write-Output 'all install checks passed'
