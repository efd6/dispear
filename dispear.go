package dispear

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"slices"
	"strings"
	"text/template"
	"unicode"
)

// DESCRIPTION sets the pipeline description field in the global context.
func DESCRIPTION(s string) { ctx.pipeline.DESCRIPTION(s) }

// VERSION sets the pipeline version field in the global context.
func VERSION(v int) { ctx.pipeline.VERSION(v) }

// METADATA sets the pipeline _meta field in the global context.
func METADATA(m map[string]any) { ctx.pipeline.METADATA(m) }

// DEPRECATED sets the pipeline deprecated status field in the global context.
func DEPRECATED(t bool) { ctx.pipeline.DEPRECATED(t) }

// ON_FAILURE sets the pipeline's global error handlers in the global context.
// Processors that are added to a pipeline's on_failure field are removed from
// the global context as independent processors.
func ON_FAILURE(h ...Renderer) { ctx.pipeline.ON_FAILURE(h...) }

// Generate emits the constructed pipeline in the global context. It should be
// called in the final line of the program. Currently the rendered pipeline is
// printed to standard output.
func Generate() error {
	err := ctx.Generate()
	if err != nil {
		panic("generate: " + err.Error())
	}
	return nil
}

var ctx = Context{tags: make(map[string][]retagger)}

// Context holds the state necessary for constructing the pipeline.
type Context struct {
	pipeline   pipeline
	processors []Renderer
	tags       map[string][]retagger
}

// Renderer is the interface required for processor rendering.
type Renderer interface {
	Render(dst io.Writer) error
}

type retagger interface {
	retag(string)
}

// Add adds a processor to the context.
func (c *Context) Add(p Renderer) {
	c.processors = append(c.processors, p)
}

func (c *Context) Generate() error {
	w := os.Stdout
	out := flag.String("out", "", "path for writing generated pipeline to (stdout if empty)")
	flag.Parse()
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	var buf bytes.Buffer
	for t, retaggers := range c.tags {
		if len(retaggers) < 2 {
			continue
		}
		w := int(math.Ceil(math.Log10(float64(len(retaggers) + 1))))
		for j, r := range retaggers {
			r.retag(fmt.Sprintf("%s_%0*d", t, w, j+1))
		}
	}
	for _, p := range c.processors {
		err := p.Render(&buf)
		if err != nil {
			return err
		}
	}
	pipelineTemplate := template.Must(template.New("pipeline").Funcs(template.FuncMap{
		"yaml":        yamlValue,
		"yaml_string": yamlString,
		"render": func(r Renderer) (string, error) {
			var buf bytes.Buffer
			err := r.Render(&buf)
			if err != nil {
				return "", err
			}
			return indent(strings.TrimSpace(buf.String()), 2), nil
		},
	}).Parse(`{{$procs := .processors}}{{with .pipeline -}}
---
{{with .Description}}description: {{yaml_string .}}
{{end -}}
{{with .Version}}version: {{.}}
{{end -}}
{{with .Metadata}}{{yaml 0 2 "_meta" .}}
{{end -}}
{{with .Deprecated}}deprecated: {{.}}
{{end -}}
processors:
{{- $procs -}}
{{- with .ErrorHandler}}
on_failure:{{range .}}
{{render .}}{{end}}
{{- end -}}
{{end}}
`))
	return pipelineTemplate.Execute(w, map[string]any{
		"pipeline":   c.pipeline,
		"processors": indent(buf.String(), 2),
	})
}

var templateHelpers = template.FuncMap{
	"comment": func(s string) string {
		return "# " + strings.Join(strings.Split(s, "\n"), "\n# ")
	},
	"gutter":      gutter,
	"yaml":        yamlValue,
	"yaml_string": yamlString,
	"render": func(r Renderer) (string, error) {
		var buf bytes.Buffer
		err := r.Render(&buf)
		if err != nil {
			return "", err
		}
		return indent(strings.TrimSpace(buf.String()), 6), nil
	},
}

type pipeline struct {
	Description  *string
	Version      *int
	Metadata     map[string]any
	Deprecated   *bool
	ErrorHandler []Renderer
}

func (pipe *pipeline) DESCRIPTION(s string) {
	if pipe.Description != nil {
		panic("multiple DESCRIPTION calls")
	}
	pipe.Description = &s
}

func (pipe *pipeline) VERSION(v int) {
	if pipe.Version != nil {
		panic("multiple VERSION calls")
	}
	pipe.Version = &v
}

func (pipe *pipeline) METADATA(m map[string]any) {
	if pipe.Metadata != nil {
		panic("multiple METADATA calls")
	}
	pipe.Metadata = m
}

func (pipe *pipeline) DEPRECATED(t bool) {
	if pipe.Deprecated != nil {
		panic("multiple DEPRECATED calls")
	}
	pipe.Deprecated = &t
}

func (pipe *pipeline) ON_FAILURE(h ...Renderer) {
	if pipe.ErrorHandler != nil {
		panic("multiple ON_ERROR calls")
	}
	pipe.ErrorHandler = h
	for i := range h {
		ctx.processors = slices.DeleteFunc(ctx.processors, func(e Renderer) bool {
			return h[i] == e
		})
	}
}

// BLANK is a no-op processor that adds a blank line to the pipeline syntax.
func BLANK() *Blank {
	b := &Blank{}
	ctx.Add(b)
	return b
}

type Blank struct {
	Comment *string
}

// COMMENT adds a comment to the blank line.
func (b *Blank) COMMENT(s string) *Blank {
	if b.Comment != nil {
		panic("multiple COMMENT calls")
	}
	b.Comment = &s
	return b
}

func (p *Blank) Render(dst io.Writer) error {
	var err error
	if p.Comment != nil {
		text := "\n# " + strings.Join(strings.Split(*p.Comment, "\n"), "\n# ")
		_, err = dst.Write([]byte(text))
		if err != nil {
			return err
		}
	} else {
		_, err = dst.Write([]byte{'\n'})
	}
	return err
}

// Look on my works, ye Mighty, and despair!
type shared[P Renderer] struct {
	parent P

	Comment       *string
	Description   *string
	Tag           string
	tagCalled     bool
	Condition     *string
	IgnoreFailure *bool
	ErrorHandler  []Renderer

	file string
	line int
}

const (
	preamble = `
{{- with .Description}}
    description: {{yaml_string .}}
{{- end -}}
{{- with .Tag}}
    tag: {{.}}
{{- end -}}
{{- with .Condition}}
{{gutter . | yaml 4 2 "if"}}
{{- end}}`

	postamble = `
{{- with .IgnoreFailure}}
    ignore_failure: {{.}}
{{- end -}}
{{- with .ErrorHandler}}
    on_failure:{{range .}}
{{render .}}{{end}}
{{- end -}}`
)

func (p *shared[P]) recDecl() {
	var ok bool
	_, p.file, p.line, ok = runtime.Caller(2)
	if !ok {
		panic("cannot get caller")
	}
}

// DESCRIPTION sets the description field of the processor.
func (p *shared[P]) DESCRIPTION(s string) P {
	if p.Description != nil {
		panic("multiple DESCRIPTION calls")
	}
	p.Description = &s
	return p.parent
}

// COMMENT adds a, potentially multi-line, comment before the processor.
func (p *shared[P]) COMMENT(s string) P {
	if p.Comment != nil {
		panic("multiple COMMENT calls")
	}
	p.Comment = &s
	return p.parent
}

// TAG sets the tag field of the processor. If the tag is not unique in the
// pipeline, it will have a numeric suffix added to make it unique.
func (p *shared[P]) TAG(s string) P {
	if p.tagCalled {
		panic("multiple TAG calls")
	}
	p.tagCalled = true
	s = PathCleaner.Replace(s)
	if s == p.Tag {
		return p.parent
	}
	if s != "" {
		ctx.tags[p.Tag] = slices.DeleteFunc(ctx.tags[p.Tag], func(e retagger) bool {
			return e == p
		})
		p.Tag = s
		ctx.tags[s] = append(ctx.tags[s], p)
	}
	return p.parent
}

func (p *shared[P]) retag(s string) {
	p.Tag = s
}

// IF sets the if condition of the processor.
func (p *shared[P]) IF(s string) P {
	if p.Condition != nil {
		panic("multiple IF calls")
	}
	if !strings.ContainsRune(s, '\n') {
		s = strings.TrimSpace(s)
	}
	p.Condition = &s
	return p.parent
}

// IGNORE_FAILURE sets the ignore_failure field to t for the processor.
func (p *shared[P]) IGNORE_FAILURE(t bool) P {
	if p.IgnoreFailure != nil {
		panic("multiple ALLOW_DUPLICATES calls")
	}
	p.IgnoreFailure = &t
	return p.parent
}

// ON_ERROR sets the on_failure field to the processors in h. Processors that
// are added to a processor's on_failure field are removed from the global
// context.
func (p *shared[P]) ON_ERROR(h ...Renderer) P {
	if p.ErrorHandler != nil {
		panic("multiple ON_ERROR calls")
	}
	p.ErrorHandler = h
	for i := range h {
		ctx.processors = slices.DeleteFunc(ctx.processors, func(e Renderer) bool {
			return h[i] == e
		})
	}
	return p.parent
}

func indent(s string, n int) string {
	ws := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if len(l) == 0 {
			continue
		}
		lines[i] = ws + l
	}
	return strings.Join(lines, "\n")
}

// This is a nasty hack to get around differential behaviour between foreach and
// other processor accepting fields.
func dedent(s string, n int) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if len(l) == 0 {
			continue
		}
		lines[i] = l[n:]
	}
	return strings.Join(lines, "\n")
}

func gutter(s string) (string, error) {
	if strings.TrimSpace(s) == "" {
		return "", errors.New("no source text")
	}
	lines := strings.Split(s, "\n")
	indented := -1
	var blanks int
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			if indented < 0 {
				lines[i] = ""
			}
			continue
		}
		if indented < 0 {
			blanks = i
			indented = strings.IndexFunc(l, func(r rune) bool { return !unicode.IsSpace(r) })
		}
		lines[i] = l[min(indented, len(l)):]
	}
	return strings.TrimRightFunc(strings.Join(lines[blanks:], "\n"), unicode.IsSpace), nil
}

// PathCleaner is applied to tags before they are used.
var PathCleaner = strings.NewReplacer(".", "_", " ", "_")

type Definition struct {
	Name, Pattern string
}

type InOutMapping struct {
	Input, Output string
}

type FieldMapping struct {
	Document, Model string
}
