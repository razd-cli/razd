# Test mise sync formatting
$projectDir = "c:\Users\dealenx\dev\razd-cli\razd\examples\nodejs-project"
cd $projectDir

# Remove mise section from Razdfile
$razdContent = @"
version: '3'

tasks:
  default:
    desc: Set up and start Node.js project
    cmds:
    - mise install
    - npm install
    - npm run dev

  install:
    desc: Install dependencies
    cmds:
    - mise install
    - npm install

  dev:
    desc: Start development server
    cmds:
    - npm run dev

  build:
    desc: Build project
    cmds:
    - npm run build

  test:
    desc: Run tests
    cmds:
    - npm test
"@

$razdContent | Out-File -Encoding UTF8 "Razdfile.yml"

# Clear tracking
Remove-Item "$env:LOCALAPPDATA\razd\tracking\*" -Force -ErrorAction SilentlyContinue

# Run once with --no-sync to establish baseline
& "c:\Users\dealenx\dev\razd-cli\razd\target\release\razd.exe" --no-sync 2>&1 | Out-Null

# Modify mise.toml timestamp
(Get-Item "mise.toml").LastWriteTime = Get-Date

# Run with sync (auto-approve via env var)
$env:RAZD_AUTO_APPROVE = "1"
& "c:\Users\dealenx\dev\razd-cli\razd\target\release\razd.exe" 2>&1 | Select-Object -First 15

Write-Host "`n=== Formatted Razdfile.yml ===`n"
Get-Content "Razdfile.yml"
