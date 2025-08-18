// Copy monaco-editor AMD assets to public for local loading
import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const src = path.join(root, 'node_modules/monaco-editor/min/vs')
const dest = path.join(root, 'public/monaco/vs')

function copyDir(srcDir, destDir) {
  if (!fs.existsSync(destDir)) fs.mkdirSync(destDir, { recursive: true })
  const entries = fs.readdirSync(srcDir, { withFileTypes: true })
  for (const entry of entries) {
    const srcPath = path.join(srcDir, entry.name)
    const destPath = path.join(destDir, entry.name)
    if (entry.isDirectory()) {
      copyDir(srcPath, destPath)
    } else if (entry.isFile()) {
      fs.copyFileSync(srcPath, destPath)
    }
  }
}

try {
  if (!fs.existsSync(src)) {
    console.error('monaco-editor assets not found at', src)
    process.exit(0)
  }
  copyDir(src, dest)
  console.log('✅ Monaco assets copied to', dest)
} catch (e) {
  console.error('Failed to copy Monaco assets:', e)
  process.exit(1)
}
