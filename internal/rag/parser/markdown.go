package parser

import (
	"bufio"
	"io"
	"strings"
)

// MarkdownParser Markdown 解析器
type MarkdownParser struct{}

// Parse 解析 Markdown 文档
func (p *MarkdownParser) Parse(file io.Reader) (*ParsedDocument, error) {
	scanner := bufio.NewScanner(file)

	var content strings.Builder
	var title string
	metadata := make(map[string]interface{})
	var sections []Section

	var currentSection *Section
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// 提取标题
		if strings.HasPrefix(line, "# ") && title == "" {
			title = strings.TrimPrefix(line, "# ")
			continue
		}

		// 检测章节
		if strings.HasPrefix(line, "## ") {
			if currentSection != nil {
				sections = append(sections, *currentSection)
			}
			currentSection = &Section{
				Title:   strings.TrimPrefix(line, "## "),
				Level:   2,
				StartLine: lineNum,
			}
		}

		if currentSection != nil {
			currentSection.Content += line + "\n"
		}

		content.WriteString(line)
		content.WriteString("\n")
	}

	if currentSection != nil {
		sections = append(sections, *currentSection)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &ParsedDocument{
		Title:    title,
		Content:  content.String(),
		Metadata: metadata,
		Sections: sections,
	}, nil
}

// SupportedFormats 支持的格式
func (p *MarkdownParser) SupportedFormats() []string {
	return []string{".md", ".markdown"}
}

// Section 文档章节
type Section struct {
	Title     string
	Level     int
	Content   string
	StartLine int
}
