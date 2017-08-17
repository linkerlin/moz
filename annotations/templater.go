package annotations

import (
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/influx6/faux/fmtwriter"
	"github.com/influx6/moz"
	"github.com/influx6/moz/ast"
	"github.com/influx6/moz/gen"
)

var (
	_ = moz.RegisterAnnotation("templaterTypesFor", TemplaterTypesForAnnotationGenerator)
	_ = moz.RegisterAnnotation("templaterTypesFor", TemplaterStructTypesForAnnotationGenerator)
	_ = moz.RegisterAnnotation("templaterTypesFor", TemplaterPackageTypesForAnnotationGenerator)
	_ = moz.RegisterAnnotation("templaterTypesFor", TemplaterInterfaceTypesForAnnotationGenerator)
)

// TypeMap defines a map type of a giving series of key-value pairs.
type TypeMap map[string]string

// Get returns the value associated with the giving key.
func (t TypeMap) Get(key string) string {
	return t[key]
}

// TemplaterStructTypesForAnnotationGenerator defines a struct level annotation generator which builds a go package in
// root of the package by using the content it receives from the annotation has a template for its output.
// package.
// Templater provides access to typenames by providing a "sel" function that gives you access to all
// arguments provided by the associated Annotation "templaterForTypes", which provides description of
// the filename, and the types to be used to replace the generic placeholders.
//
// Annotation: @templaterTypesFor
//
// Example:
// 1. Create a template that uses the "Go" generator, identified with the id "Mob" which will
// generate template for all types by using a template from a @templater with id of "Mob", define
// @templater anywhere either in package, struct, type or interface level.
//
// @templater(id => Mob, gen => Go, {
//
//   func Add(m {{sel TYPE1}}, n {{sel TYPE2}}) {{sel TYPE3}} {
//
//   }
//
// })
//
// 2. Add @templaterTypesFor annotation on any level (Type, Struct, Interface, Package) to have the code
// generated from the details provided.
//
// @templaterTypesFor(id => Mob, filename => bob_gen.go, TYPE1 => int32, TYPE2 => int32, TYPE3 => int64)
// @templaterTypesFor(id => Mob, filename => bib_gen.go, TYPE1 => int, TYPE2 => int, TYPE3 => int64)
//
func TemplaterStructTypesForAnnotationGenerator(toDir string, an ast.AnnotationDeclaration, ty ast.StructDeclaration, pkg ast.PackageDeclaration) ([]gen.WriteDirective, error) {
	templaterId, ok := an.Params["id"]
	if !ok {
		return nil, errors.New("No templater id provided")
	}

	// Get all templaters AnnotationDeclaration.
	templaters := pkg.AnnotationsFor("templater")

	var targetTemplater ast.AnnotationDeclaration

	// Search for templater with associated ID, if not found, return error, if multiple found, use the first.
	for _, targetTemplater = range templaters {
		if targetTemplater.Params["id"] != templaterId {
			continue
		}

		break
	}

	if targetTemplater.Template == "" {
		return nil, errors.New("Expected Template from annotation")
	}

	var directives []gen.WriteDirective

	genName := strings.ToLower(targetTemplater.Params["gen"])
	genID := strings.ToLower(targetTemplater.Params["id"])

	fileName, ok := an.Params["filename"]
	if !ok {
		fileName = fmt.Sprintf("%s_templater_types_for_gen.%s", genID, genName)
	}

	typeGen := gen.Block(gen.SourceTextWith(targetTemplater.Template, template.FuncMap{
		"sel": TypeMap(an.Params).Get,
	}, struct {
		TemplateParams     TypeMap
		TemplateForParams  TypeMap
		TypeForAnnotation  ast.AnnotationDeclaration
		TemplateAnnotation ast.AnnotationDeclaration
		StructDeclr        ast.StructDeclaration
		Package            ast.PackageDeclaration
	}{
		StructDeclr:        ty,
		Package:            pkg,
		TypeForAnnotation:  an,
		TemplateAnnotation: targetTemplater,
		TemplateParams:     TypeMap(targetTemplater.Params),
		TemplateForParams:  TypeMap(an.Params),
	}))

	switch genName {
	case "partial.go":

		pkgGen := gen.Block(
			gen.Commentary(
				gen.Text("Autogenerated using the moz templater annotation."),
			),
			gen.Package(
				gen.Name(pkg.Package),
				typeGen,
			),
		)

		directives = append(directives, gen.WriteDirective{
			FileName:     fileName,
			DontOverride: true,
			Writer:       fmtwriter.New(pkgGen, true, true),
		})

	case "go":
		directives = append(directives, gen.WriteDirective{
			FileName:     fileName,
			DontOverride: true,

			Writer: fmtwriter.New(typeGen, true, true),
		})

	default:
		directives = append(directives, gen.WriteDirective{
			Writer:       typeGen,
			DontOverride: true,
			FileName:     fileName,
		})
	}

	return directives, nil
}

// TemplaterInterfaceTypesForAnnotationGenerator defines a package level annotation generator which builds a go package in
// root of the package by using the content it receives from the annotation has a template for its output.
// package.
// Templater provides access to typenames by providing a "sel" function that gives you access to all
// arguments provided by the associated Annotation "templaterForTypes", which provides description of
// the filename, and the types to be used to replace the generic placeholders.
//
// Annotation: @templaterTypesFor
//
// Example:
// 1. Create a template that uses the "Go" generator, identified with the id "Mob" which will
// generate template for all types by using a template from a @templater with id of "Mob", define
// @templater anywhere either in package, struct, type or interface level.
//
// @templater(id => Mob, gen => Go, {
//
//   func Add(m {{sel TYPE1}}, n {{sel TYPE2}}) {{sel TYPE3}} {
//
//   }
//
// })
//
// 2. Add @templaterTypesFor annotation on any level (Type, Struct, Interface, Package) to have the code
// generated from the details provided.
//
// @templaterTypesFor(id => Mob, filename => bob_gen.go, TYPE1 => int32, TYPE2 => int32, TYPE3 => int64)
// @templaterTypesFor(id => Mob, filename => bib_gen.go, TYPE1 => int, TYPE2 => int, TYPE3 => int64)
//
func TemplaterInterfaceTypesForAnnotationGenerator(toDir string, an ast.AnnotationDeclaration, ty ast.InterfaceDeclaration, pkg ast.PackageDeclaration) ([]gen.WriteDirective, error) {
	templaterId, ok := an.Params["id"]
	if !ok {
		return nil, errors.New("No templater id provided")
	}

	// Get all templaters AnnotationDeclaration.
	templaters := pkg.AnnotationsFor("templater")

	var targetTemplater ast.AnnotationDeclaration

	// Search for templater with associated ID, if not found, return error, if multiple found, use the first.
	for _, targetTemplater = range templaters {
		if targetTemplater.Params["id"] != templaterId {
			continue
		}

		break
	}

	if targetTemplater.Template == "" {
		return nil, errors.New("Expected Template from annotation")
	}

	var directives []gen.WriteDirective

	genName := strings.ToLower(targetTemplater.Params["gen"])
	genID := strings.ToLower(targetTemplater.Params["id"])

	fileName, ok := an.Params["filename"]
	if !ok {
		fileName = fmt.Sprintf("%s_templater_types_for_gen.%s", genID, genName)
	}

	typeGen := gen.Block(gen.SourceTextWith(targetTemplater.Template, template.FuncMap{
		"sel": TypeMap(an.Params).Get,
	}, struct {
		TemplateParams     TypeMap
		TemplateForParams  TypeMap
		TypeForAnnotation  ast.AnnotationDeclaration
		TemplateAnnotation ast.AnnotationDeclaration
		InterfaceDeclr     ast.InterfaceDeclaration
		Package            ast.PackageDeclaration
	}{
		InterfaceDeclr:     ty,
		Package:            pkg,
		TypeForAnnotation:  an,
		TemplateAnnotation: targetTemplater,
		TemplateParams:     TypeMap(targetTemplater.Params),
		TemplateForParams:  TypeMap(an.Params),
	}))

	switch genName {
	case "partial.go":

		pkgGen := gen.Block(
			gen.Commentary(
				gen.Text("Autogenerated using the moz templater annotation."),
			),
			gen.Package(
				gen.Name(ast.WhichPackage(toDir, pkg)),
				typeGen,
			),
		)

		directives = append(directives, gen.WriteDirective{
			FileName:     fileName,
			DontOverride: true,

			Writer: fmtwriter.New(pkgGen, true, true),
		})

	case "go":
		directives = append(directives, gen.WriteDirective{
			FileName:     fileName,
			DontOverride: true,
			Writer:       fmtwriter.New(typeGen, true, true),
		})

	default:
		directives = append(directives, gen.WriteDirective{
			Writer:       typeGen,
			DontOverride: true,
			FileName:     fileName,
		})
	}

	return directives, nil
}

// TemplaterPackageTypesForAnnotationGenerator defines a package level annotation generator which builds a go package in
// root of the package by using the content it receives from the annotation has a template for its output.
// package.
// Templater provides access to typenames by providing a "sel" function that gives you access to all
// arguments provided by the associated Annotation "templaterForTypes", which provides description of
// the filename, and the types to be used to replace the generic placeholders.
//
// Annotation: @templaterTypesFor
//
// Example:
// 1. Create a template that uses the "Go" generator, identified with the id "Mob" which will
// generate template for all types by using a template from a @templater with id of "Mob", define
// @templater anywhere either in package, struct, type or interface level.
//
// @templater(id => Mob, gen => Go, {
//
//   func Add(m {{sel TYPE1}}, n {{sel TYPE2}}) {{sel TYPE3}} {
//
//   }
//
// })
//
// 2. Add @templaterTypesFor annotation on any level (Type, Struct, Interface, Package) to have the code
// generated from the details provided.
//
// @templaterTypesFor(id => Mob, filename => bob_gen.go, TYPE1 => int32, TYPE2 => int32, TYPE3 => int64)
// @templaterTypesFor(id => Mob, filename => bib_gen.go, TYPE1 => int, TYPE2 => int, TYPE3 => int64)
//
func TemplaterPackageTypesForAnnotationGenerator(toDir string, an ast.AnnotationDeclaration, pkg ast.PackageDeclaration) ([]gen.WriteDirective, error) {
	templaterId, ok := an.Params["id"]
	if !ok {
		return nil, errors.New("No templater id provided")
	}

	// Get all templaters AnnotationDeclaration.
	templaters := pkg.AnnotationsFor("templater")

	var targetTemplater ast.AnnotationDeclaration

	// Search for templater with associated ID, if not found, return error, if multiple found, use the first.
	for _, targetTemplater = range templaters {
		if targetTemplater.Params["id"] != templaterId {
			continue
		}

		break
	}

	if targetTemplater.Template == "" {
		return nil, errors.New("Expected Template from annotation")
	}

	var directives []gen.WriteDirective

	genName := strings.ToLower(targetTemplater.Params["gen"])
	genID := strings.ToLower(targetTemplater.Params["id"])

	fileName, ok := an.Params["filename"]
	if !ok {
		fileName = fmt.Sprintf("%s_templater_types_for_gen.%s", genID, genName)
	}

	typeGen := gen.Block(gen.SourceTextWith(targetTemplater.Template, template.FuncMap{
		"sel": TypeMap(an.Params).Get,
	}, struct {
		TemplateParams     TypeMap
		TemplateForParams  TypeMap
		TypeForAnnotation  ast.AnnotationDeclaration
		TemplateAnnotation ast.AnnotationDeclaration
		Package            ast.PackageDeclaration
	}{
		Package:            pkg,
		TypeForAnnotation:  an,
		TemplateAnnotation: targetTemplater,
		TemplateParams:     TypeMap(targetTemplater.Params),
		TemplateForParams:  TypeMap(an.Params),
	}))

	switch genName {
	case "partial.go":

		pkgGen := gen.Block(
			gen.Commentary(
				gen.Text("Autogenerated using the moz templater annotation."),
			),
			gen.Package(
				gen.Name(pkg.Package),
				typeGen,
			),
		)

		directives = append(directives, gen.WriteDirective{
			FileName:     fileName,
			DontOverride: true,
			Writer:       fmtwriter.New(pkgGen, true, true),
		})

	case "go":
		directives = append(directives, gen.WriteDirective{
			FileName:     fileName,
			DontOverride: true,

			Writer: fmtwriter.New(typeGen, true, true),
		})

	default:
		directives = append(directives, gen.WriteDirective{
			Writer:       typeGen,
			DontOverride: true,
			FileName:     fileName,
		})
	}

	return directives, nil
}

// TemplaterTypesForAnnotationGenerator defines a package level annotation generator which builds a go package in
// root of the package by using the content it receives from the annotation has a template for its output.
// package.
// Templater provides access to typenames by providing a "sel" function that gives you access to all
// arguments provided by the associated Annotation "templaterForTypes", which provides description of
// the filename, and the types to be used to replace the generic placeholders.
//
// Annotation: @templaterTypesFor
//
// Example:
// 1. Create a template that uses the "Go" generator, identified with the id "Mob" which will
// generate template for all types by using a template from a @templater with id of "Mob", define
// @templater anywhere either in package, struct, type or interface level.
//
// @templater(id => Mob, gen => Go, {
//
//   func Add(m {{sel TYPE1}}, n {{sel TYPE2}}) {{sel TYPE3}} {
//
//   }
//
// })
//
// 2. Add @templaterTypesFor annotation on any level (Type, Struct, Interface, Package) to have the code
// generated from the details provided.
//
// @templaterTypesFor(id => Mob, filename => bob_gen.go, TYPE1 => int32, TYPE2 => int32, TYPE3 => int64)
// @templaterTypesFor(id => Mob, filename => bib_gen.go, TYPE1 => int, TYPE2 => int, TYPE3 => int64)
//
func TemplaterTypesForAnnotationGenerator(toDir string, an ast.AnnotationDeclaration, ty ast.TypeDeclaration, pkg ast.PackageDeclaration) ([]gen.WriteDirective, error) {
	templaterId, ok := an.Params["id"]
	if !ok {
		return nil, errors.New("No templater id provided")
	}

	// Get all templaters AnnotationDeclaration.
	templaters := pkg.AnnotationsFor("templater")

	var targetTemplater ast.AnnotationDeclaration

	// Search for templater with associated ID, if not found, return error, if multiple found, use the first.
	for _, targetTemplater = range templaters {
		if targetTemplater.Params["id"] != templaterId {
			continue
		}

		break
	}

	if targetTemplater.Template == "" {
		return nil, errors.New("Expected Template from annotation")
	}

	var directives []gen.WriteDirective

	genName := strings.ToLower(targetTemplater.Params["gen"])
	genID := strings.ToLower(targetTemplater.Params["id"])

	fileName, ok := an.Params["filename"]
	if !ok {
		fileName = fmt.Sprintf("%s_templater_types_for_gen.%s", genID, genName)
	}

	typeGen := gen.Block(gen.SourceTextWith(targetTemplater.Template, template.FuncMap{
		"sel": TypeMap(an.Params).Get,
	}, struct {
		TemplateParams     TypeMap
		TemplateForParams  TypeMap
		TypeForAnnotation  ast.AnnotationDeclaration
		TemplateAnnotation ast.AnnotationDeclaration
		TypeDeclr          ast.TypeDeclaration
		Package            ast.PackageDeclaration
	}{
		TypeDeclr:          ty,
		Package:            pkg,
		TypeForAnnotation:  an,
		TemplateAnnotation: targetTemplater,
		TemplateParams:     TypeMap(targetTemplater.Params),
		TemplateForParams:  TypeMap(an.Params),
	}))

	switch genName {
	case "partial.go":

		pkgGen := gen.Block(
			gen.Commentary(
				gen.Text("Autogenerated using the moz templater annotation."),
			),
			gen.Package(
				gen.Name(pkg.Package),
				typeGen,
			),
		)

		directives = append(directives, gen.WriteDirective{
			FileName:     fileName,
			DontOverride: true,
			Writer:       fmtwriter.New(pkgGen, true, true),
		})

	case "go":
		directives = append(directives, gen.WriteDirective{
			FileName:     fileName,
			DontOverride: true,

			Writer: fmtwriter.New(typeGen, true, true),
		})

	default:
		directives = append(directives, gen.WriteDirective{
			Writer:       typeGen,
			DontOverride: true,
			FileName:     fileName,
		})
	}

	return directives, nil
}
