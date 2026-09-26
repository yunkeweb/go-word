# Silent Microsoft Word compatibility check for test_output_docs/*.docx.
# Starts a headless Word.Application (DisplayAlerts = 0) and opens each file
# with OpenNoRepairDialog so an OpenXML repair would surface as an exception
# instead of a dialog.
param(
    [string]$Dir = "test_output_docs"
)

$ErrorActionPreference = "Stop"
$wdDoNotSaveChanges = 0
$wdAlertsNone = 0
$msoAutomationSecurityForceDisable = 3

$root = (Get-Location).Path
$dirPath = Join-Path $root $Dir
if (-not (Test-Path $dirPath)) {
    Write-Error "directory not found: $dirPath"
    exit 2
}

$files = @(Get-ChildItem -Path $dirPath -Filter "*.docx" | Sort-Object Name)
if ($files.Count -eq 0) {
    Write-Error "no .docx in $dirPath"
    exit 2
}

Write-Host ("MS Word COM check: {0} files in {1}" -f $files.Count, $dirPath)
Write-Host ("{0,-42}  {1,-6}  {2}" -f "FILE", "WORD", "DETAIL")

$word = $null
$pass = 0
$fail = 0
$results = New-Object System.Collections.Generic.List[string]

try {
    $word = New-Object -ComObject Word.Application
    $word.Visible = $false
    $word.DisplayAlerts = $wdAlertsNone
    $word.ScreenUpdating = $false
    try { $word.AutomationSecurity = $msoAutomationSecurityForceDisable } catch { }
    try { $word.FeatureInstall = 0 } catch { }

    $missing = [Type]::Missing
    $hasNoRepair = $null -ne ($word.Documents | Get-Member -Name OpenNoRepairDialog -MemberType Method -ErrorAction SilentlyContinue)

    foreach ($f in $files) {
        $doc = $null
        $status = "PASS"
        $detail = ""
        try {
            $full = $f.FullName
            if ($hasNoRepair) {
                $doc = $word.Documents.OpenNoRepairDialog(
                    $full, $false, $true, $false,
                    $missing, $missing, $true
                )
            } else {
                $doc = $word.Documents.Open(
                    $full, $false, $true, $false,
                    $missing, $missing, $true
                )
            }
            if ($null -eq $doc) {
                throw "Documents.Open returned null"
            }
            $chars = 0
            try { $chars = $doc.Content.Characters.Count } catch { $chars = -1 }
            $detail = "chars=$chars"
            if ($doc.ReadOnlyRecommended) { $detail += " readonlyRecommended" }
        } catch {
            $status = "FAIL"
            $detail = $_.Exception.Message
            if ($detail.Length -gt 180) { $detail = $detail.Substring(0, 180) }
        } finally {
            if ($null -ne $doc) {
                try { $doc.Close([ref]$wdDoNotSaveChanges) } catch {
                    try { $doc.Close($wdDoNotSaveChanges) } catch { }
                }
                try { [void][System.Runtime.InteropServices.Marshal]::ReleaseComObject($doc) } catch { }
            }
        }

        if ($status -eq "PASS") { $pass++ } else { $fail++ }
        $line = ("{0,-42}  {1,-6}  {2}" -f $f.Name, $status, $detail)
        Write-Host $line
        $results.Add(("{0}`t{1}`t{2}" -f $f.Name, $status, $detail))
    }
} catch {
    Write-Host ("WORD_ENGINE  FAIL  cannot start Word.Application: " + $_.Exception.Message)
    $fail = $files.Count
    $pass = 0
} finally {
    if ($null -ne $word) {
        try { $word.Quit($wdDoNotSaveChanges) } catch {
            try { $word.Quit() } catch { }
        }
        try { [void][System.Runtime.InteropServices.Marshal]::ReleaseComObject($word) } catch { }
        $word = $null
        [GC]::Collect()
        [GC]::WaitForPendingFinalizers()
    }
}

$report = Join-Path $dirPath "word_com_report.tsv"
@(
    "file`tword`tdetail"
    $results
) | Set-Content -Path $report -Encoding UTF8

Write-Host ""
Write-Host ("MS Word totals: {0} PASS  {1} FAIL  ({2} files)" -f $pass, $fail, $files.Count)
Write-Host ("report: {0}" -f $report)

if ($fail -gt 0) { exit 1 }
exit 0
