# Script to normalize all line endings to LF (Linux standard)
# Run this on Windows to convert CRLF to LF

Write-Host "Normalizing line endings to LF..." -ForegroundColor Green

# Get all text files
$files = git ls-files

foreach ($file in $files) {
    # Skip binary files
    if ($file -match '\.(exe|dll|so|dylib|png|jpg|jpeg|gif|ico|zip|gz|7z)$') {
        continue
    }
    
    # Read file content
    $content = Get-Content $file -Raw
    
    if ($content) {
        # Replace CRLF with LF
        $content = $content -replace "`r`n", "`n"
        
        # Write back without BOM
        $utf8NoBom = New-Object System.Text.UTF8Encoding $false
        [System.IO.File]::WriteAllText($file, $content, $utf8NoBom)
        
        Write-Host "  Normalized: $file" -ForegroundColor Gray
    }
}

Write-Host "`nDone! All files now use LF line endings." -ForegroundColor Green
Write-Host "Run 'git status' to see changes." -ForegroundColor Yellow
