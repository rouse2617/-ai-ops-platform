package chunker

import (
	"strings"
)

// ChunkConfig 分块配置
type ChunkConfig struct {
	ChunkSize    int      // Token 数量
	ChunkOverlap int      // 重叠 Token
	Separators   []string // 分隔符列表
}

// RecursiveCharacterSplitter 递归字符分割器
type RecursiveCharacterSplitter struct {
	config ChunkConfig
}

// NewRecursiveCharacterSplitter 创建分割器
func NewRecursiveCharacterSplitter(config ChunkConfig) *RecursiveCharacterSplitter {
	if len(config.Separators) == 0 {
		config.Separators = []string{"\n\n", "\n", ". ", " "}
	}
	return &RecursiveCharacterSplitter{config: config}
}

// Split 分割文本
func (s *RecursiveCharacterSplitter) Split(text string) []string {
	return s.splitRecursive(text, s.config.Separators)
}

// splitRecursive 递归分割
func (s *RecursiveCharacterSplitter) splitRecursive(text string, separators []string) []string {
	if len(text) <= s.config.ChunkSize {
		return []string{text}
	}

	if len(separators) == 0 {
		return s.splitBySize(text)
	}

	separator := separators[0]
	parts := strings.Split(text, separator)

	var chunks []string
	var currentChunk strings.Builder

	for _, part := range parts {
		if currentChunk.Len()+len(part)+len(separator) <= s.config.ChunkSize {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString(separator)
			}
			currentChunk.WriteString(part)
		} else {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
			}

			if len(part) > s.config.ChunkSize {
				subChunks := s.splitRecursive(part, separators[1:])
				chunks = append(chunks, subChunks...)
			} else {
				currentChunk.WriteString(part)
			}
		}
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return s.addOverlap(chunks)
}

// splitBySize 按大小强制分割
func (s *RecursiveCharacterSplitter) splitBySize(text string) []string {
	var chunks []string
	for i := 0; i < len(text); i += s.config.ChunkSize {
		end := i + s.config.ChunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}

// addOverlap 添加重叠
func (s *RecursiveCharacterSplitter) addOverlap(chunks []string) []string {
	if s.config.ChunkOverlap == 0 || len(chunks) <= 1 {
		return chunks
	}

	overlapped := make([]string, len(chunks))
	overlapped[0] = chunks[0]

	for i := 1; i < len(chunks); i++ {
		prev := chunks[i-1]
		curr := chunks[i]

		overlapStart := len(prev) - s.config.ChunkOverlap
		if overlapStart < 0 {
			overlapStart = 0
		}

		overlap := prev[overlapStart:]
		overlapped[i] = overlap + curr
	}

	return overlapped
}
