package setting

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rakyll/statik/fs"
	"github.com/teamlint/ardan/pkg"
	_ "github.com/teamlint/ardan/res"
)

type TmplType = string

const (
	TmplTypeOrigin TmplType = ".orig"
	TmplTypeCode   TmplType = ".tmpl"
	TmplTypeView   TmplType = ".tmpl"
	TmplTypeSample TmplType = ".samp"
	TmplTypeBuild  TmplType = ".ardan"
)

type Directive = string

const (
	// go files
	GoMainFileTmpl = "main.go.tmpl"
	GoMainFile     = "main.go"
	GoGenFileTmpl  = "gen.go.tmpl"
	GoFileSuffix   = ".go"
	// template
	InternalTmplDir = "templates"
	SyncDir         = "sync"
	// directive
	DirectiveSync Directive = "ardan:sync"
	DirectiveGen  Directive = "ardan:gen"
)

type Setting struct {
	Template   *template.Template
	FileSystem http.FileSystem
	Origins    []string // origin files
	Layouts    []string // project layout dir names
	Codes      []string // source code names
	Samples    []string // sample names
	DBDriver   string   // database driver name
	DBName     string   // database name
	DBConnStr  string   // database connection string
	GoMod      string
	// layout
	InternalTmplDir string // internal template dir
	TmplDir         string // template dir
	Output          string // output root dir
	Cmd             string // cmd dir
	Sync            string // sync dir
	Doc             string // documents root directory
	App             string // application dir
	Model           string // domain layer directory
	Service         string // service layer directory
	Repository      string // repository layer directory
	Server          string // server layer directory
	ServerModule    string // server module directory
	ServerGlobal    string // server global directory
	Controller      string // controller directory
	Handler         string // handler directory
	Middleware      string // middleware directory
	Sample          bool
}

// Options setting options
type Options struct {
	TmplDir   string
	OutputDir string
	FuncMap   template.FuncMap
	DBDriver  string // database driver name
	DBName    string // database name
	DBConnStr string // database connection string
	GoModName string
	Config    string // config file
	// project layout
	CmdDir          string // command root directory
	DocDir          string // documents root directory
	AppDir          string // application layer directory
	ModelDir        string // domain layer directory
	ServiceDir      string // service layer directory
	RepositoryDir   string // repository layer directory
	ServerDir       string // server layer directory
	ServerModuleDir string // server module directory
	ServerGlobalDir string // server global directory
	ControllerDir   string // controller directory
	HandlerDir      string // handler directory
	MiddlewareDir   string // middleware directory
	Sample          bool
}

// New init settings
func New(opt Options) *Setting {
	tmplDir := clean(opt.TmplDir)
	outputDir := clean(opt.OutputDir)
	cmdDir := clean(opt.CmdDir)
	docDir := clean(opt.DocDir)
	appDir := clean(opt.AppDir)
	modelDir := clean(opt.ModelDir)
	serviceDir := clean(opt.ServiceDir)
	repositoryDir := clean(opt.RepositoryDir)
	serverDir := clean(opt.ServerDir)
	serverModuleDir := clean(opt.ServerModuleDir)
	serverGlobalDir := clean(opt.ServerGlobalDir)
	controllerDir := clean(opt.ControllerDir)
	handlerDir := clean(opt.HandlerDir)
	middlewareDir := clean(opt.MiddlewareDir)

	instance := &Setting{
		Origins:   make([]string, 0),
		Layouts:   make([]string, 0),
		Codes:     make([]string, 0),
		Samples:   make([]string, 0),
		DBDriver:  opt.DBDriver,
		DBName:    opt.DBName,
		DBConnStr: opt.DBConnStr,
		GoMod:     opt.GoModName,
		// internal
		InternalTmplDir: InternalTmplDir,
		Sync:            SyncDir,
		// layout
		TmplDir:      tmplDir,
		Output:       outputDir,
		Cmd:          cmdDir,
		Doc:          docDir,
		App:          appDir,
		Model:        modelDir,
		Service:      serviceDir,
		Repository:   repositoryDir,
		Server:       serverDir,
		ServerModule: serverModuleDir,
		ServerGlobal: serverGlobalDir,
		Controller:   controllerDir,
		Handler:      handlerDir,
		Middleware:   middlewareDir,
		Sample:       opt.Sample,
	}

	hfs, err := fs.New()
	if err != nil {
		panic(fmt.Errorf("res.FileSystem New err=%v\n", err))
	}
	instance.FileSystem = hfs

	err = instance.walkTemplates(opt)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	return instance

}

func (s *Setting) walkTemplates(opt Options) error {
	root := template.New("")

	err := fs.Walk(s.FileSystem, "/", func(path string, info os.FileInfo, e1 error) error {
		if e1 != nil {
			return e1
		}
		// log.Printf("path=%v\n", path)
		if info.IsDir() {
			s.Layouts = append(s.Layouts, filepath.Join(opt.OutputDir, path))
			return nil
		}
		// is original file
		if strings.HasSuffix(path, TmplTypeOrigin) {
			s.Origins = append(s.Origins, path)
			return nil
		}
		// is template
		if strings.HasSuffix(path, TmplTypeCode) {
			b, e2 := fs.ReadFile(s.FileSystem, path)
			if e2 != nil {
				return fmt.Errorf("res.ReadFile file=%v err=%v\n", path, e2)
			}
			// log.Printf("res.origin.file=%v\n", string(b))
			t := root.New(path).Funcs(defaultFuncMap()).Funcs(opt.FuncMap)
			t, e2 = t.Parse(string(b))
			if e2 != nil {
				return e2
			}
			s.Codes = append(s.Codes, path)
			return nil
		}
		// is sample
		if opt.Sample && strings.HasSuffix(path, TmplTypeSample) {
			b, e2 := fs.ReadFile(s.FileSystem, path)
			if e2 != nil {
				return fmt.Errorf("res.ReadFile file=%v err=%v\n", path, e2)
			}
			// log.Printf("res.origin.file=%v\n", string(b))
			t := root.New(path).Funcs(defaultFuncMap()).Funcs(opt.FuncMap)
			t, e2 = t.Parse(string(b))
			if e2 != nil {
				return e2
			}
			s.Samples = append(s.Samples, path)
			return nil
		}
		return nil
	})

	s.Template = root
	return err
}

func defaultFuncMap() template.FuncMap {
	fm := template.FuncMap{}
	fm["clean"] = clean
	fm["randomString"] = pkg.RandomString
	return fm
}

func clean(path string) string {
	path = strings.TrimPrefix(path, ".")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	return path
}

func (s *Setting) TargetFile(srcname string) string {
	ext := filepath.Ext(srcname)

	switch ext {
	case TmplTypeOrigin, TmplTypeCode, TmplTypeSample, TmplTypeBuild:
		dst := strings.TrimSuffix(srcname, ext)
		// log.Printf("[TargetFile] dst=%v,ext=%v\n", dst, filepath.Ext(dst))
		return clean(filepath.Join(s.Output, dst))
	}
	return filepath.Join(s.Output, srcname)
}

func (s *Setting) SourceFile(srcname string) string {
	return filepath.Join(s.InternalTmplDir, srcname)
}

func (s *Setting) TmplFile(path ...string) string {
	root := []string{"/"}
	root = append(root, path...)
	return filepath.Join(root...)
}

func (s *Setting) HasPrefix(path, layout string) bool {
	name := strings.TrimPrefix(path, "/")
	return strings.HasPrefix(name, layout)
}

func (s *Setting) HasDirective(doc, directive string) (string, bool) {
	doc = strings.TrimPrefix(doc, "//")
	doc = strings.TrimPrefix(doc, " ")
	return doc, strings.HasPrefix(doc, directive)
}
func (s *Setting) IsGen(path string) bool {
	return strings.HasSuffix(path, GoGenFileTmpl)
}
