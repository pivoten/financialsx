#!/bin/bash

# Script to replace console.log statements with logger service

echo "Replacing console.log statements with logger service..."

# Find all TypeScript files
find src -name "*.tsx" -o -name "*.ts" | while read file; do
  # Skip the logger.ts file itself
  if [[ "$file" == *"logger.ts" ]]; then
    continue
  fi
  
  # Check if file has console.log statements
  if grep -q "console\." "$file"; then
    echo "Processing: $file"
    
    # Create backup
    cp "$file" "$file.bak"
    
    # Add logger import if not present
    if ! grep -q "import.*logger" "$file"; then
      # Add import after the first import statement
      sed -i '' '1,/^import/ {
        /^import/ a\
import logger from '\''../services/logger'\''
      }' "$file"
    fi
    
    # Replace console statements
    sed -i '' \
      -e 's/console\.log(/logger.debug(/g' \
      -e 's/console\.info(/logger.info(/g' \
      -e 's/console\.warn(/logger.warn(/g' \
      -e 's/console\.error(/logger.error(/g' \
      "$file"
    
    # Remove backup if successful
    if [ $? -eq 0 ]; then
      rm "$file.bak"
    else
      echo "Error processing $file, restoring backup"
      mv "$file.bak" "$file"
    fi
  fi
done

echo "Replacement complete!"
echo "Note: You may need to adjust import paths for the logger service based on file location"