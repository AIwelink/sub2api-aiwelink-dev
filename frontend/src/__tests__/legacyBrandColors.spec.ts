import { readFileSync, readdirSync } from 'node:fs'
import { extname, join, relative, resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const frontendRoot = resolve(process.cwd())
const sourceRoot = resolve(frontendRoot, 'src')
const scannedExtensions = new Set(['.css', '.js', '.ts', '.vue'])
const legacyBrand = /#14b8a6|#0d9488|#0f766e|rgba?\(20,\s*184,\s*166/gi

function sourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)

    if (entry.isDirectory()) {
      return entry.name === '__tests__' ? [] : sourceFiles(path)
    }

    return scannedExtensions.has(extname(entry.name)) ? [path] : []
  })
}

describe('legacy brand colors', () => {
  it('does not hard-code the former teal brand palette', () => {
    const files = [...sourceFiles(sourceRoot), resolve(frontendRoot, 'tailwind.config.js')]
    const matches = files.flatMap((path) => {
      const lines = readFileSync(path, 'utf8').split(/\r?\n/)

      return lines.flatMap((line, index) => {
        legacyBrand.lastIndex = 0
        return legacyBrand.test(line)
          ? [`${relative(frontendRoot, path)}:${index + 1}: ${line.trim()}`]
          : []
      })
    })

    expect(matches, matches.join('\n')).toEqual([])
  })
})
