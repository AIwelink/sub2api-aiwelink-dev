import { readFileSync, readdirSync } from 'node:fs'
import { extname, join, relative, resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const sourceRoot = resolve(process.cwd(), 'src')
const stylePath = resolve(sourceRoot, 'style.css')
const primaryBackground = /(?:bg|from)-primary-(?:400|500|600|700)/
const whiteForeground = /text-white/

function sourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)

    if (entry.isDirectory()) {
      return entry.name === '__tests__' ? [] : sourceFiles(path)
    }

    return ['.ts', '.vue'].includes(extname(entry.name)) ? [path] : []
  })
}

function quotedStrings(source: string, quotes = '["\'`]'): string[] {
  return Array.from(source.matchAll(new RegExp(`(${quotes})([\\s\\S]*?)\\1`, 'g')), (match) => match[2])
}

function vueClassGroups(source: string): string[] {
  const staticClasses = Array.from(
    source.matchAll(/(?<!:)class\s*=\s*(["'])([\s\S]*?)\1/g),
    (match) => match[2]
  )
  const boundClasses = Array.from(
    source.matchAll(/:class\s*=\s*(["'])([\s\S]*?)\1/g),
    (match) => quotedStrings(match[2], match[1] === '"' ? "['`]" : '["`]')
  ).flat()

  return [...staticClasses, ...boundClasses]
}

function classGroups(path: string, source: string): string[] {
  return extname(path) === '.vue' ? vueClassGroups(source) : quotedStrings(source)
}

describe('primary foreground contrast', () => {
  it('does not pair primary backgrounds with white foregrounds', () => {
    const offenders = sourceFiles(sourceRoot).flatMap((path) => {
      const source = readFileSync(path, 'utf8')

      return classGroups(path, source)
        .filter((classes) => primaryBackground.test(classes) && whiteForeground.test(classes))
        .map((classes) => `${relative(sourceRoot, path)}: ${classes.trim().replace(/\s+/g, ' ')}`)
    })

    expect(offenders, offenders.join('\n')).toEqual([])
  })

  it('uses the semantic primary foreground in the shared button', () => {
    const styles = readFileSync(stylePath, 'utf8')
    const buttonRule = styles.match(/\.btn-primary\s*\{([\s\S]*?)\}/)?.[1] ?? ''

    expect(buttonRule).toContain('text-on-primary')
    expect(buttonRule).not.toContain('text-white')
  })
})
