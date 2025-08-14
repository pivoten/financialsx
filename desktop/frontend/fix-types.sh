#!/bin/bash

# Script to fix common TypeScript errors

echo "Starting TypeScript fixes..."

# Fix all event handler parameters
echo "Fixing event handler parameters..."
find src -name "*.tsx" -o -name "*.ts" | while read file; do
  # Fix onChange handlers
  sed -i '' 's/onChange={(e)/onChange={(e: React.ChangeEvent<HTMLInputElement>)/g' "$file"
  sed -i '' 's/onChange={e/onChange={(e: React.ChangeEvent<HTMLInputElement>)/g' "$file"
  sed -i '' 's/onSubmit={(e)/onSubmit={(e: React.FormEvent<HTMLFormElement>)/g' "$file"
  sed -i '' 's/onClick={()/onClick={(): void =>/g' "$file"
  
  # Fix parameter types
  sed -i '' 's/\([(,]\)\s*\([a-zA-Z_][a-zA-Z0-9_]*\))\s*=>/\1\2: any) =>/g' "$file"
done

echo "Fixing array types..."
find src -name "*.tsx" -o -name "*.ts" | while read file; do
  # Fix useState arrays
  sed -i '' 's/useState(\[\])/useState<any[]>([])/g' "$file"
  sed -i '' 's/useState(null)/useState<any>(null)/g' "$file"
done

echo "Done!"