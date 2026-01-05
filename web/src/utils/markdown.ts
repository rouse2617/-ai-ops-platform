/**
 * Markdown Rendering Utilities
 * 
 * Optimized markdown rendering with caching and performance improvements.
 */

import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github-dark.css'

// Configure marked once
marked.setOptions({
  breaks: true,
  gfm: true
})

// Custom renderer for code highlighting
const renderer = new marked.Renderer()
renderer.code = function(code: string, language: string | undefined) {
  if (language && hljs.getLanguage(language)) {
    try {
      const highlighted = hljs.highlight(code, { language }).value
      return `<pre><code class="hljs language-${language}">${highlighted}</code></pre>`
    } catch {
      // Fall through
    }
  }
  const highlighted = hljs.highlightAuto(code).value
  return `<pre><code class="hljs">${highlighted}</code></pre>`
}
marked.use({ renderer })

// Simple LRU cache for rendered markdown
const markdownCache = new Map<string, string>()
const MAX_CACHE_SIZE = 100

function getCacheKey(content: string): string {
  // Simple hash function for cache keys
  let hash = 0
  for (let i = 0; i < content.length; i++) {
    const char = content.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash = hash & hash // Convert to 32bit integer
  }
  return hash.toString(36)
}

/**
 * Render markdown to HTML with caching
 */
export function renderMarkdown(content: string): string {
  const cacheKey = getCacheKey(content)
  
  if (markdownCache.has(cacheKey)) {
    return markdownCache.get(cacheKey)!
  }

  const html = marked(content) as string

  // Manage cache size
  if (markdownCache.size >= MAX_CACHE_SIZE) {
    const firstKey = markdownCache.keys().next().value as string
    if (firstKey) markdownCache.delete(firstKey)
  }

  markdownCache.set(cacheKey, html)
  return html
}

/**
 * Clear markdown cache
 */
export function clearMarkdownCache(): void {
  markdownCache.clear()
}

/**
 * Extract plain text from markdown
 */
export function markdownToPlainText(content: string): string {
  return content
    .replace(/```[\s\S]*?```/g, '') // Remove code blocks
    .replace(/`[^`]+`/g, '') // Remove inline code
    .replace(/\*\*/g, '') // Remove bold
    .replace(/\*/g, '') // Remove italic
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1') // Replace links with text
    .replace(/\n+/g, ' ') // Replace newlines with space
    .trim()
}
