package imports2jsonl

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"io"
	"iter"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type RawToken struct{ token.Token }

func (t RawToken) String() string {
	return t.Token.String()
}

type CommentDto struct {
	Slash int    `json:"slash"`
	Text  string `json:"text"`
}

type CommentGroupDto struct {
	List []CommentDto `json:"list"`
}

type IdentDto struct {
	Name string `json:"name"`
}

type PositionDto struct {
	Filename string `json:"filename"`
	Offset   int    `json:"offset"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

type RawPosition token.Position

func (r RawPosition) Dto() PositionDto {
	return PositionDto(r)
}

type BasicLitDto struct {
	ValuePos PositionDto `json:"value_pos"`
	Kind     string      `json:"kind"`
	Value    string      `json:"value"`
}

type ImportSpecDto struct {
	Doc     CommentGroupDto `json:"doc"`
	Name    IdentDto        `json:"name"`
	Path    BasicLitDto     `json:"path"`
	Comment CommentGroupDto `json:"comment"`
}

type RawComment struct{ *ast.Comment }

func (c RawComment) Text() string {
	if nil == c.Comment {
		return ""
	}

	return c.Comment.Text
}

func (c RawComment) Slash() int {
	if nil == c.Comment {
		return 0
	}

	return int(c.Comment.Slash)
}

func (c RawComment) Dto() CommentDto {
	return CommentDto{
		Slash: c.Slash(),
		Text:  c.Text(),
	}
}

type RawCommentGroup struct{ *ast.CommentGroup }

func (g RawCommentGroup) Text() string {
	if g.CommentGroup == nil {
		return ""
	}
	return g.CommentGroup.Text()
}

func (g RawCommentGroup) List() []*ast.Comment {
	if nil == g.CommentGroup {
		return nil
	}
	return g.CommentGroup.List
}

type RawIdent struct{ *ast.Ident }

func (i RawIdent) Name() string {
	if i.Ident == nil {
		return ""
	}
	return i.Ident.Name
}

type RawBasicLit struct{ *ast.BasicLit }

func (b RawBasicLit) ValuePos() int {
	if b.BasicLit == nil {
		return 0
	}
	return int(b.BasicLit.ValuePos)
}

func (b RawBasicLit) Kind() string {
	if b.BasicLit == nil {
		return ""
	}
	return b.BasicLit.Kind.String()
}

func (b RawBasicLit) Value() string {
	if b.BasicLit == nil {
		return ""
	}
	return b.BasicLit.Value
}

type RawImportSpec struct{ *ast.ImportSpec }

func (i RawImportSpec) Doc() RawCommentGroup {
	if nil == i.ImportSpec {
		return RawCommentGroup{}
	}

	return RawCommentGroup{CommentGroup: i.ImportSpec.Doc}
}

func (i RawImportSpec) Name() RawIdent {
	if i.ImportSpec == nil {
		return RawIdent{}
	}
	return RawIdent{Ident: i.ImportSpec.Name}
}

func (i RawImportSpec) Path() RawBasicLit {
	if i.ImportSpec == nil {
		return RawBasicLit{}
	}
	return RawBasicLit{BasicLit: i.ImportSpec.Path}
}

func (i RawImportSpec) Comment() RawCommentGroup {
	if i.ImportSpec == nil {
		return RawCommentGroup{}
	}
	return RawCommentGroup{CommentGroup: i.ImportSpec.Comment}
}

func (i RawImportSpec) EndPos() int {
	if i.ImportSpec == nil {
		return 0
	}
	return int(i.ImportSpec.EndPos)
}

func (g RawCommentGroup) Dto() CommentGroupDto {
	if g.CommentGroup == nil {
		return CommentGroupDto{}
	}

	list := make([]CommentDto, 0, len(g.CommentGroup.List))
	for _, c := range g.CommentGroup.List {
		list = append(list, RawComment{c}.Dto())
	}
	return CommentGroupDto{List: list}
}

func (i RawIdent) Dto() IdentDto {
	if i.Ident == nil {
		return IdentDto{}
	}
	return IdentDto{Name: i.Ident.Name}
}

func (b RawBasicLit) Dto(fset *token.FileSet) BasicLitDto {
	if b.BasicLit == nil {
		return BasicLitDto{}
	}
	return BasicLitDto{
		ValuePos: RawPosition(fset.Position(b.BasicLit.ValuePos)).Dto(),
		Kind:     b.BasicLit.Kind.String(),
		Value:    strings.Trim(b.BasicLit.Value, `"`),
	}
}

func (i RawImportSpec) Dto(fset *token.FileSet) ImportSpecDto {
	if i.ImportSpec == nil {
		return ImportSpecDto{}
	}

	return ImportSpecDto{
		Doc:     i.Doc().Dto(),
		Name:    i.Name().Dto(),
		Path:    i.Path().Dto(fset),
		Comment: i.Comment().Dto(),
	}
}

type RawFile struct{ *ast.File }

func (f RawFile) Imports() []*ast.ImportSpec {
	if nil == f.File {
		return nil
	}

	return f.File.Imports
}

func (f RawFile) Name(fset *token.FileSet) string {
	return fset.Position(f.File.Package).Filename
}

type Iter[T any] iter.Seq[T]

func (i Iter[T]) Raw() iter.Seq[T] { return iter.Seq[T](i) }

func IterMap[T, U any](original Iter[T], mapper func(T) U) Iter[U] {
	return func(yield func(U) bool) {
		for item := range original {
			var mapd U = mapper(item)
			if !yield(mapd) {
				return
			}
		}
	}
}

type FileDto struct {
	Name    string
	Imports []ImportSpecDto
}

func (f RawFile) Dto(fset *token.FileSet) FileDto {
	return FileDto{
		Name: f.Name(fset),
		Imports: slices.Collect(
			IterMap(
				Iter[*ast.ImportSpec](slices.Values(f.Imports())),
				func(original *ast.ImportSpec) ImportSpecDto {
					return RawImportSpec{original}.Dto(fset)
				},
			).Raw(),
		),
	}
}

func (f FileDto) ToJSON(enc *json.Encoder) error {
	for _, imp := range f.Imports {
		e := enc.Encode(imp)
		if nil != e {
			return e
		}
	}
	return nil
}

type RawPass struct{ *analysis.Pass }

func (p RawPass) Files() []*ast.File {
	if nil == p.Pass {
		return nil
	}

	return p.Pass.Files
}

type PassDto struct {
	Files []FileDto
}

func (p RawPass) Dto() PassDto {
	return PassDto{
		Files: slices.Collect(
			IterMap(
				Iter[*ast.File](slices.Values(p.Files())),
				func(original *ast.File) FileDto {
					return RawFile{original}.Dto(p.Pass.Fset)
				},
			).Raw(),
		),
	}
}

func (p PassDto) ToJSON(enc *json.Encoder) error {
	for _, file := range p.Files {
		e := file.ToJSON(enc)
		if nil != e {
			return e
		}
	}
	return nil
}

type Config struct {
	*json.Encoder

	Name string
	Doc  string
}

func (c Config) Run(pass *analysis.Pass) (any, error) {
	raw := RawPass{pass}
	var dto PassDto = raw.Dto()
	return nil, dto.ToJSON(c.Encoder)
}

func (c Config) WithWriter(w io.Writer) Config {
	c.Encoder = json.NewEncoder(w)
	return c
}

func (c Config) ToAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: c.Name,
		Doc:  c.Doc,
		Run:  c.Run,
	}
}

const NameDefault string = "import_details2jsonl"
const DocDefault string = "Import detail extractor"

var WriterDefault io.Writer = io.Discard //nolint:gochecknoglobals

//nolint:gochecknoglobals
var ConfigDefault Config = Config{
	Encoder: json.NewEncoder(WriterDefault),
	Name:    NameDefault,
	Doc:     DocDefault,
}
