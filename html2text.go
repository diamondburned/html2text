package html2text

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
	"github.com/olekukonko/tablewriter"
	"github.com/ssor/bom"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Options provide toggles and overrides to control specific rendering behaviors.
type Options struct {
	PrettyTables        bool                 // Turns on pretty ASCII rendering for table elements.
	PrettyTablesOptions *PrettyTablesOptions // Configures pretty ASCII rendering for table elements.
	OmitLinks           bool                 // Turns on omitting links
	TextOnly            bool                 // Returns only plain text
}

// PrettyTablesOptions overrides tablewriter behaviors
type PrettyTablesOptions struct {
	AutoFormatHeader     bool
	AutoWrapText         bool
	ReflowDuringAutoWrap bool
	ColWidth             int
	ColumnSeparator      string
	RowSeparator         string
	CenterSeparator      string
	HeaderAlignment      int
	FooterAlignment      int
	Alignment            int
	ColumnAlignment      []int
	NewLine              string
	HeaderLine           bool
	RowLine              bool
	AutoMergeCells       bool
	Borders              tablewriter.Border
}

// NewPrettyTablesOptions creates PrettyTablesOptions with default settings
func NewPrettyTablesOptions() *PrettyTablesOptions {
	return &PrettyTablesOptions{
		AutoFormatHeader:     true,
		AutoWrapText:         true,
		ReflowDuringAutoWrap: true,
		ColWidth:             tablewriter.MAX_ROW_WIDTH,
		ColumnSeparator:      tablewriter.COLUMN,
		RowSeparator:         tablewriter.ROW,
		CenterSeparator:      tablewriter.CENTER,
		HeaderAlignment:      tablewriter.ALIGN_DEFAULT,
		FooterAlignment:      tablewriter.ALIGN_DEFAULT,
		Alignment:            tablewriter.ALIGN_DEFAULT,
		ColumnAlignment:      []int{},
		NewLine:              tablewriter.NEWLINE,
		HeaderLine:           true,
		RowLine:              false,
		AutoMergeCells:       false,
		Borders:              tablewriter.Border{Left: true, Right: true, Bottom: true, Top: true},
	}
}

// FromHTMLNode renders text output from a pre-parsed HTML document.
func FromHTMLNode(doc *html.Node, o ...Options) (string, error) {
	var options Options
	if len(o) > 0 {
		options = o[0]
	}

	ctx := textifyTraverseContext{
		options: options,
	}
	ctx.lineWrapper = lineWrapper{
		out:   &ctx.buf,
		width: 78,
	}

	if err := ctx.traverse(doc); err != nil {
		return "", err
	}

	text := strings.TrimSpace(newlineRe.ReplaceAllString(
		strings.Replace(ctx.buf.String(), "\n ", "\n", -1), "\n\n"),
	)
	return text, nil
}

// FromReader renders text output after parsing HTML for the specified
// io.Reader.
func FromReader(reader io.Reader, options ...Options) (string, error) {
	newReader, err := bom.NewReaderWithoutBom(reader)
	if err != nil {
		return "", err
	}
	doc, err := html.Parse(newReader)
	if err != nil {
		return "", err
	}
	return FromHTMLNode(doc, options...)
}

// FromString parses HTML from the input string, then renders the text form.
func FromString(input string, options ...Options) (string, error) {
	bs := bom.CleanBom([]byte(input))
	text, err := FromReader(bytes.NewReader(bs), options...)
	if err != nil {
		return "", err
	}
	return text, nil
}

var (
	newlineRe = regexp.MustCompile(`\n\n+`)
)

// traverseTableCtx holds text-related context.
type textifyTraverseContext struct {
	buf strings.Builder

	prefix          string
	tableCtx        tableTraverseContext
	options         Options
	endsWithSpace   bool
	justClosedDiv   bool
	blockquoteLevel int
	tableLevel      int
	lineWrapper     lineWrapper
	isPre           bool
}

// tableTraverseContext holds table ASCII-form related context.
type tableTraverseContext struct {
	header     []string
	body       [][]string
	footer     []string
	tmpRow     int
	isInFooter bool
}

func (tableCtx *tableTraverseContext) init() {
	tableCtx.body = [][]string{}
	tableCtx.header = []string{}
	tableCtx.footer = []string{}
	tableCtx.isInFooter = false
	tableCtx.tmpRow = 0
}

func (ctx *textifyTraverseContext) sub() *textifyTraverseContext {
	subCtx := textifyTraverseContext{}
	subCtx.options = ctx.options
	subCtx.lineWrapper = lineWrapper{
		out:   &subCtx.buf,
		width: ctx.lineWrapper.width,
	}
	return &subCtx
}

func (ctx *textifyTraverseContext) handleElement(node *html.Node) error {
	ctx.justClosedDiv = false

	switch node.DataAtom {
	case atom.Br:
		return ctx.emit("\n\n")

	case atom.H1, atom.H2, atom.H3:
		subCtx := ctx.sub()
		if err := subCtx.traverseChildren(node); err != nil {
			return err
		}

		str := subCtx.buf.String()
		if ctx.options.TextOnly {
			return ctx.emit(str + "\n\n")
		}

		dividerLen := 0
		for _, line := range strings.Split(str, "\n") {
			line = strings.TrimRightFunc(line, unicode.IsSpace)
			if lineLen := runewidth.StringWidth(line); lineLen > dividerLen {
				dividerLen = lineLen
			}
		}

		var divider string
		if node.DataAtom == atom.H1 {
			divider = strings.Repeat("*", dividerLen)
		} else {
			divider = strings.Repeat("-", dividerLen)
		}

		ctx.emit("\n\n")

		if node.DataAtom == atom.H1 {
			ctx.emit(divider)
			ctx.emit("\n")
		}

		ctx.emit(str)
		ctx.emit("\n")

		ctx.emit(divider)
		ctx.emit("\n\n")
		return nil

	case atom.Blockquote:
		ctx.blockquoteLevel++
		if !ctx.options.TextOnly {
			ctx.prefix = strings.Repeat(">", ctx.blockquoteLevel) + " "
		}
		if err := ctx.emit("\n"); err != nil {
			return err
		}
		if ctx.blockquoteLevel == 1 {
			if err := ctx.emit("\n"); err != nil {
				return err
			}
		}
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}
		ctx.blockquoteLevel--
		if !ctx.options.TextOnly {
			ctx.prefix = strings.Repeat(">", ctx.blockquoteLevel)
		}
		if ctx.blockquoteLevel > 0 {
			ctx.prefix += " "
		}
		return ctx.emit("\n\n")

	case atom.Div:
		ctx.lineWrapper.flush()
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}
		var err error
		if !ctx.justClosedDiv {
			err = ctx.emit("\n")
		}
		ctx.justClosedDiv = true
		return err

	case atom.Li:
		if !ctx.options.TextOnly {
			if err := ctx.emit("- "); err != nil {
				return err
			}
		}

		if err := ctx.traverseChildren(node); err != nil {
			return err
		}

		return ctx.emit("\n")

	case atom.B, atom.Strong:
		subCtx := ctx.sub()
		subCtx.endsWithSpace = true
		if err := subCtx.traverseChildren(node); err != nil {
			return err
		}
		str := subCtx.buf.String()
		if ctx.options.TextOnly {
			return ctx.emit(str)
		}
		return ctx.emit("*" + str + "*")

	case atom.A:
		linkText := ""
		// For simple link element content with single text node only, peek at the link text.
		if node.FirstChild != nil && node.FirstChild.NextSibling == nil && node.FirstChild.Type == html.TextNode {
			linkText = node.FirstChild.Data
		}

		// If image is the only child, take its alt text as the link text.
		if img := node.FirstChild; img != nil && node.LastChild == img && img.DataAtom == atom.Img {
			if altText := getAttrVal(img, "alt"); altText != "" {
				if err := ctx.emit(altText); err != nil {
					return err
				}
			}
		} else if err := ctx.traverseChildren(node); err != nil {
			return err
		}

		hrefLink := ""
		if attrVal := getAttrVal(node, "href"); attrVal != "" {
			attrVal = ctx.normalizeHrefLink(attrVal)
			// Don't print link href if it matches link element content or if the link is empty.
			if (!ctx.options.OmitLinks && attrVal != "" && linkText != attrVal) || !ctx.options.TextOnly {
				hrefLink = "(" + attrVal + ")"
			}
		}

		return ctx.emit(hrefLink)

	case atom.P, atom.Ul:
		return ctx.paragraphHandler(node)

	case atom.Table:
		ctx.tableLevel++
		defer func() { ctx.tableLevel-- }()

		fallthrough
	case atom.Tfoot, atom.Th, atom.Tr, atom.Td:
		if ctx.options.PrettyTables {
			return ctx.handleTableElement(node)
		} else if node.DataAtom == atom.Table {
			return ctx.paragraphHandler(node)
		}
		return ctx.traverseChildren(node)

	case atom.Pre:
		ctx.isPre = true
		err := ctx.traverseChildren(node)
		ctx.isPre = false
		return err

	case atom.Style, atom.Script, atom.Head:
		// Ignore the subtree.
		return nil

	default:
		return ctx.traverseChildren(node)
	}
}

// paragraphHandler renders node children surrounded by double newlines.
func (ctx *textifyTraverseContext) paragraphHandler(node *html.Node) error {
	if err := ctx.emit("\n\n"); err != nil {
		return err
	}
	if err := ctx.traverseChildren(node); err != nil {
		return err
	}
	return ctx.emit("\n\n")
}

// handleTableElement is only to be invoked when options.PrettyTables is active.
func (ctx *textifyTraverseContext) handleTableElement(node *html.Node) error {
	if !ctx.options.PrettyTables {
		panic("handleTableElement invoked when PrettyTables not active")
	}

	switch node.DataAtom {
	case atom.Table:
		if err := ctx.emit("\n\n"); err != nil {
			return err
		}

		// Re-intialize all table context.
		ctx.tableCtx.init()

		// Browse children, enriching context with table data.
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}

		buf := &bytes.Buffer{}
		table := tablewriter.NewWriter(buf)
		if ctx.options.PrettyTablesOptions != nil {
			options := ctx.options.PrettyTablesOptions
			table.SetAutoFormatHeaders(options.AutoFormatHeader)
			table.SetAutoWrapText(options.AutoWrapText)
			table.SetReflowDuringAutoWrap(options.ReflowDuringAutoWrap)
			table.SetColWidth(options.ColWidth)
			table.SetColumnSeparator(options.ColumnSeparator)
			table.SetRowSeparator(options.RowSeparator)
			table.SetCenterSeparator(options.CenterSeparator)
			table.SetHeaderAlignment(options.HeaderAlignment)
			table.SetFooterAlignment(options.FooterAlignment)
			table.SetAlignment(options.Alignment)
			table.SetColumnAlignment(options.ColumnAlignment)
			table.SetNewLine(options.NewLine)
			table.SetHeaderLine(options.HeaderLine)
			table.SetRowLine(options.RowLine)
			table.SetAutoMergeCells(options.AutoMergeCells)
			table.SetBorders(options.Borders)
		}
		table.SetHeader(ctx.tableCtx.header)
		table.SetFooter(ctx.tableCtx.footer)
		table.AppendBulk(ctx.tableCtx.body)

		// Render the table using ASCII.
		table.Render()
		if err := ctx.emit(buf.String()); err != nil {
			return err
		}

		return ctx.emit("\n\n")

	case atom.Tfoot:
		ctx.tableCtx.isInFooter = true
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}
		ctx.tableCtx.isInFooter = false

	case atom.Tr:
		ctx.tableCtx.body = append(ctx.tableCtx.body, []string{})
		if err := ctx.traverseChildren(node); err != nil {
			return err
		}
		ctx.tableCtx.tmpRow++

	case atom.Th:
		res, err := ctx.renderEachChild(node)
		if err != nil {
			return err
		}

		ctx.tableCtx.header = append(ctx.tableCtx.header, res)

	case atom.Td:
		res, err := ctx.renderEachChild(node)
		if err != nil {
			return err
		}

		if ctx.tableCtx.isInFooter {
			ctx.tableCtx.footer = append(ctx.tableCtx.footer, res)
		} else {
			ctx.tableCtx.body[ctx.tableCtx.tmpRow] = append(ctx.tableCtx.body[ctx.tableCtx.tmpRow], res)
		}

	}
	return nil
}

func (ctx *textifyTraverseContext) traverse(node *html.Node) error {
	switch node.Type {
	default:
		return ctx.traverseChildren(node)

	case html.TextNode:
		var data string
		if ctx.isPre {
			data = node.Data
		} else {
			data = strings.TrimSpace(node.Data)
		}
		return ctx.emit(data)

	case html.ElementNode:
		return ctx.handleElement(node)
	}
}

func (ctx *textifyTraverseContext) traverseChildren(node *html.Node) error {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if err := ctx.traverse(c); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *textifyTraverseContext) emit(data string) error {
	if ctx.tableLevel > 0 {
		ctx.lineWrapper.flush()
		ctx.buf.WriteString(data)
		return nil
	}

	switch data {
	case "":
		return nil
	case "\n":
		ctx.lineWrapper.flush()
		return nil
	case "\n\n":
		ctx.lineWrapper.flushN(2)
		return nil
	}

	ctx.lineWrapper.write(data)
	return nil
}

func (ctx *textifyTraverseContext) normalizeHrefLink(link string) string {
	link = strings.TrimSpace(link)
	link = strings.TrimPrefix(link, "mailto:")
	return link
}

// renderEachChild visits each direct child of a node and collects the sequence of
// textuual representaitons separated by a single newline.
func (ctx *textifyTraverseContext) renderEachChild(node *html.Node) (string, error) {
	buf := &bytes.Buffer{}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		s, err := FromHTMLNode(c, ctx.options)
		if err != nil {
			return "", err
		}
		if _, err = buf.WriteString(s); err != nil {
			return "", err
		}
		if c.NextSibling != nil {
			if err = buf.WriteByte('\n'); err != nil {
				return "", err
			}
		}
	}
	return buf.String(), nil
}

func getAttrVal(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

// lineWrapper is copied from package go/doc. It is slightly modified to support
// proper rune widths.
type lineWrapper struct {
	out       io.Writer
	width     int
	n         int
	nl        int
	pendSpace int
	printed   bool
}

var nl = []byte("\n")
var space = []byte(" ")

func (l *lineWrapper) write(text string) {
	if l.n == 0 && l.printed {
		l.flush() // blank line before new paragraph
	}

	l.printed = true
	l.nl = 0

	for _, f := range strings.Fields(text) {
		w := runewidth.StringWidth(f)
		// wrap if line is too long
		if l.n > 0 && l.n+l.pendSpace+w > l.width {
			l.out.Write(nl)
			l.n = 0
			l.pendSpace = 0
		}
		l.out.Write(space[:l.pendSpace])
		l.out.Write([]byte(f))
		l.n += l.pendSpace + w
		l.pendSpace = 1
	}
}

func (l *lineWrapper) flush() {
	l.flushN(1)
}

func (l *lineWrapper) flushN(n int) {
	if l.n == 0 && l.nl >= n {
		return
	}

	n -= l.nl
	if n < 1 {
		return
	}

	if n == 1 {
		l.out.Write(nl)
	} else {
		l.out.Write(bytes.Repeat(nl, n))
	}

	l.pendSpace = 0
	l.n = 0
	l.nl += n
}
