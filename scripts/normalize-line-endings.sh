#!/bin/bash
# Script to normalize all line endings to LF (Linux standard)

echo "Normalizing line endings to LF..."

# Get all text files tracked by git
git ls-files | while read -r file; do
    # Skip binary files
    if [[ "$file" =~ \.(exe|dll|so|dylib|png|jpg|jpeg|gif|ico|zip|gz|7z)$ ]]; then
        continue
    fi
    
    # Convert CRLF to LF using dos2unix or sed
    if command -v dos2unix &> /dev/null; then
        dos2unix "$file" 2>/dev/null
    else
        # Fallback to sed if dos2unix not available
        sed -i 's/\r$//' "$file" 2>/dev/null
    fi
    
    echo "  Normalized: $file"
done

echo ""
echo "Done! All files now use LF line endings."
echo "Run 'git status' to see changes."
