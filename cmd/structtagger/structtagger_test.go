package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func Test_tagValueGetter(t *testing.T) {
	tests := []struct {
		name   string
		tag    reflect.StructTag
		tagName string
		want   string
		wantOk bool
	}{
		{
			name:    "json compound",
			tag:     `protobuf:"bytes,1,opt,name=contentLeft,proto3" json:"content_left,omitempty"`,
			tagName: "json",
			want:    "content_left",
			wantOk:  true,
		},
		{
			name:    "json simple",
			tag:     `json:"content_left"`,
			tagName: "json",
			want:    "content_left",
			wantOk:  true,
		},
		{
			name:    "protobuf",
			tag:     `protobuf:"bytes,1,opt,name=contentLeft,proto3" json:"content_left,omitempty"`,
			tagName: "protobuf",
			want:    "contentLeft",
			wantOk:  true,
		},
		{
			name:    "bson",
			tag:     `json:"id,omitempty" bson:"_id"`,
			tagName: "bson",
			want:    "_id",
			wantOk:  true,
		},
		{
			name:    "yaml",
			tag:     `yaml:"foo_bar,omitempty"`,
			tagName: "yaml",
			want:    "foo_bar",
			wantOk:  true,
		},
		{
			name:    "toml",
			tag:     `toml:"baz_qux"`,
			tagName: "toml",
			want:    "baz_qux",
			wantOk:  true,
		},
		{
			name:    "redis",
			tag:     `json:"id,omitempty" redis:"id,tag"`,
			tagName: "redis",
			want:    "id",
			wantOk:  true,
		},
		{
			name:    "redis untagged",
			tag:     `json:"id,omitempty"`,
			tagName: "redis",
			want:    "",
			wantOk:  false,
		},
		{
			name:    "undefined tag",
			tag:     `json:"id,omitempty" bson:"_id"`,
			tagName: "gson",
			want:    "",
			wantOk:  false,
		},
		{
			name:    "defined yet unknown tag",
			tag:     `cooltag:"the coolness"`,
			tagName: "cooltag",
			want:    "",
			wantOk:  false,
		},
		{
			name:    "protobuf empty name",
			tag:     `protobuf:"bytes,1,opt,name=,proto3"`,
			tagName: "protobuf",
			want:    "",
			wantOk:  true,
		},
		{
			name:    "protobuf malformed missing name",
			tag:     `protobuf:"bytes,1,opt,proto3"`,
			tagName: "protobuf",
			want:    "",
			wantOk:  false,
		},
		{
			name:    "json empty",
			tag:     `json:",omitempty"`,
			tagName: "json",
			want:    "",
			wantOk:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := tagValueGetter(tt.tag, tt.tagName)
			if got != tt.want {
				t.Errorf("tagValueGetter() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("tagValueGetter() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_validateTag(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"json", true},
		{"bson", true},
		{"yaml", true},
		{"toml", true},
		{"protobuf", true},
		{"redis", true},
		{"xml", false},
		{"", false},
		{"JSON", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateTag(tt.name); got != tt.want {
				t.Errorf("validateTag(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_isDirectory(t *testing.T) {
	t.Run("directory", func(t *testing.T) {
		dir := t.TempDir()
		if !isDirectory(dir) {
			t.Errorf("isDirectory(%q) = false, want true", dir)
		}
	})
	t.Run("file", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "foo.txt")
		if err := os.WriteFile(f, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}
		if isDirectory(f) {
			t.Errorf("isDirectory(%q) = true, want false", f)
		}
	})
}

func Test_isPackage(t *testing.T) {
	t.Run("current directory", func(t *testing.T) {
		if !isPackage(".") {
			t.Error("isPackage(.) = false, want true")
		}
	})
}

func loadTempPackage(t *testing.T, src string) *packages.Package {
	t.Helper()
	dir := t.TempDir()
	mod := "module testpkg\ngo 1.23\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "source.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedImports | packages.NeedDeps,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if len(pkgs[0].Errors) > 0 {
		t.Fatalf("package errors: %v", pkgs[0].Errors)
	}
	return pkgs[0]
}

func TestGenerator_generate(t *testing.T) {
	t.Run("basic json", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make  string `+"`"+`json:"make"`+"`"+`
	Model string `+"`"+`json:"model,omitempty"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json"})

		want := `
const (
	CarMakeJson = "make"
	CarModelJson = "model"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("unexported skipped", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make      string `+"`"+`json:"make"`+"`"+`
	unexported string `+"`"+`json:"unexported"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json"})

		want := `
const (
	CarMakeJson = "make"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("underscore skipped", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make string `+"`"+`json:"make"`+"`"+`
	_    string `+"`"+`json:"ignored"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json"})

		want := `
const (
	CarMakeJson = "make"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("XXX_ skipped", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make        string `+"`"+`json:"make"`+"`"+`
	XXX_Private string `+"`"+`json:"private"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json"})

		want := `
const (
	CarMakeJson = "make"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("dash tag skipped", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make string `+"`"+`json:"make"`+"`"+`
	Year int    `+"`"+`json:"-"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json"})

		want := `
const (
	CarMakeJson = "make"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("missing tag fallback", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make  string `+"`"+`bson:"make_bson"`+"`"+`
	Model string `+"`"+`json:"model"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json"})

		want := `
const (
	CarMakeJson = "Make"
	CarModelJson = "model"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("multiple tags", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make string `+"`"+`json:"make" bson:"make_bson"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json", "bson"})

		want := `
const (
	CarMakeJson = "make"
	CarMakeBson = "make_bson"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("redis present and missing mixed", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make  string `+"`"+`json:"make" redis:"redis_make"`+"`"+`
	Model string `+"`"+`json:"model"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json", "redis"})

		want := `
const (
	CarMakeJson = "make"
	CarMakeRedis = "redis_make"
	CarModelJson = "model"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("protobuf tag", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make string `+"`"+`protobuf:"bytes,1,opt,name=make,proto3" json:"make,omitempty"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"protobuf"})

		want := `
const (
	CarMakeProtobuf = "make"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("multiple types", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Make string `+"`"+`json:"make"`+"`"+`
}

type Bike struct {
	Brand string `+"`"+`json:"brand"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json"})
		g.generate("Bike", []string{"json"})

		want := `
const (
	CarMakeJson = "make"
)

const (
	BikeBrandJson = "brand"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("sorting", func(t *testing.T) {
		pkg := loadTempPackage(t, `package testpkg

type Car struct {
	Zebra string `+"`"+`json:"zebra"`+"`"+`
	Apple string `+"`"+`json:"apple"`+"`"+`
	Mango string `+"`"+`json:"mango"`+"`"+`
}`)
		g := &Generator{}
		g.addPackage(pkg)
		g.generate("Car", []string{"json"})

		want := `
const (
	CarAppleJson = "apple"
	CarMangoJson = "mango"
	CarZebraJson = "zebra"
)
`
		if got := g.buf.String(); got != want {
			t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
		}
	})

}

func TestGenerator_format(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		g := &Generator{}
		g.Printf("package foo\n\nconst (\n\tBar = \"baz\"\n)\n")
		src := g.format()
		if src == nil {
			t.Fatal("format returned nil")
		}
		if !strings.Contains(string(src), "package foo") {
			t.Error("formatted output missing package clause")
		}
	})

	t.Run("invalid go", func(t *testing.T) {
		g := &Generator{}
		g.Printf("package {broken}\n")
		src := g.format()
		if src == nil {
			t.Fatal("format returned nil")
		}
		// Should return raw bytes on error.
		if !strings.Contains(string(src), "{broken}") {
			t.Error("expected raw output on format error")
		}
	})
}
