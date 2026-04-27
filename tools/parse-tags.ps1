param([string]$Tag)

if ($Tag -match '^v(\d+)\.(\d+)\.(\d+)$') {
    $v = $Tag -replace '^v', ''
    $p = $v.Split('.')
    Write-Output "$($p[0]) $($p[0]).$($p[1]) $v"
} else {
    Write-Output $Tag
}
