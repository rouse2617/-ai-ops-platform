package parser

import "io"

// DocumentParser 文档解析器接口
type DocumentParser interface {
	Parse(file io.Reader) (*ParsedDocument, error)
	SupportedFormats() []string
}

// ParsedDocument 解析后的文档
type ParsedDocument struct {
	Title    string
	Content  string
	Metadata map[string]interface{}
	Sections []Section
}

// ParserRegistry 解析器注册表
type ParserRegistry struct {
	parsers map[string]DocumentParser
}

// NewParserRegistry 创建解析器注册表
func NewParserRegistry() *ParserRegistry {
	registry := &ParserRegistry{
		parsers: make(map[string]DocumentParser),
	}

	// 注册默认解析器
	registry.Register(&MarkdownParser{})

	return registry
}

// Register 注册解析器
func (r *ParserRegistry) Register(parser DocumentParser) {
	for _, format := range parser.SupportedFormats() {
		r.parsers[format] = parser
	}
}

// GetParser 获取解析器
func (r *ParserRegistry) GetParser(format string) (DocumentParser, bool) {
	parser, ok := r.parsers[format]
	return parser, ok
}
