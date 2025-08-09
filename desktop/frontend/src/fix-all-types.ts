#!/usr/bin/env node

// This file contains type fixes that need to be applied to all components
// Run: npx ts-node src/fix-all-types.ts

import * as fs from 'fs';
import * as path from 'path';
import { glob } from 'glob';

const fixTypeScriptErrors = async () => {
  console.log('Starting comprehensive TypeScript fixes...');
  
  // Get all TypeScript files
  const files = await glob('src/**/*.{ts,tsx}', {
    ignore: ['node_modules/**', 'dist/**']
  });
  
  for (const file of files) {
    let content = fs.readFileSync(file, 'utf8');
    let modified = false;
    
    // Fix onChange handlers
    if (content.includes('onChange={') && !content.includes('React.ChangeEvent')) {
      content = content.replace(
        /onChange=\{(\(?)e(\)?)(\s*)=>/g,
        'onChange={(e: React.ChangeEvent<HTMLInputElement>) =>'
      );
      modified = true;
    }
    
    // Fix onSubmit handlers
    if (content.includes('onSubmit={') && !content.includes('React.FormEvent')) {
      content = content.replace(
        /onSubmit=\{(\(?)e(\)?)(\s*)=>/g,
        'onSubmit={(e: React.FormEvent<HTMLFormElement>) =>'
      );
      modified = true;
    }
    
    // Fix onClick handlers without parameters
    content = content.replace(
      /onClick=\{(\(\))\s*=>/g,
      'onClick={() =>'
    );
    
    // Fix async functions with no parameter types
    content = content.replace(
      /const (\w+) = async \(([a-zA-Z_][a-zA-Z0-9_]*)\) =>/g,
      'const $1 = async ($2: any) =>'
    );
    
    // Fix useState without types
    content = content.replace(
      /useState\(\[\]\)/g,
      'useState<any[]>([])'
    );
    content = content.replace(
      /useState\(null\)/g,
      'useState<any>(null)'
    );
    content = content.replace(
      /useState\(''\)/g,
      'useState<string>("")'
    );
    content = content.replace(
      /useState\(false\)/g,
      'useState<boolean>(false)'
    );
    content = content.replace(
      /useState\(true\)/g,
      'useState<boolean>(true)'
    );
    content = content.replace(
      /useState\(0\)/g,
      'useState<number>(0)'
    );
    
    // Fix .map/.filter/.forEach without types
    content = content.replace(
      /\.(map|filter|forEach|find|some|every|reduce)\((\([a-zA-Z_][a-zA-Z0-9_]*)(,\s*[a-zA-Z_][a-zA-Z0-9_]*)?\)\s*=>/g,
      '.$1(($2: any$3) =>'
    );
    
    if (modified) {
      fs.writeFileSync(file, content);
      console.log(`Fixed: ${file}`);
    }
  }
  
  console.log('TypeScript fixes completed!');
};

fixTypeScriptErrors().catch(console.error);