package build

import (
	"fmt"
	"go/ast"
	"go/token"
)

func GenerateTemplate(file *ast.File, importPath string, isWatch bool) {
	importPathValue := fmt.Sprintf("\"%s\"", importPath)

	component := ""
	functionComponent := ""
	props := ""
	for _, decl := range file.Decls {
		if (component != "" || functionComponent != "") && props != "" {
			break
		}

		switch c := decl.(type) {
		case *ast.GenDecl:
			if t, ok := c.Specs[0].(*ast.TypeSpec); ok {
				if fieldList, ok := t.Type.(*ast.StructType); ok {
					for _, field := range fieldList.Fields.List {
						searchComponent := func(expr *ast.SelectorExpr) bool {
							if ident, ok := expr.X.(*ast.Ident); ok {
								if ident.Name == "react" {
									if expr.Sel.Name == "ComponentDef" {
										component = t.Name.Name
										field.Tag = &ast.BasicLit{Kind: token.STRING, Value: importPathValue}
										return true
									}
									if expr.Sel.Name == "FunctionComponent" {
										functionComponent = t.Name.Name
										field.Tag = &ast.BasicLit{Kind: token.STRING, Value: importPathValue}
										return true
									}
									if expr.Sel.Name == "Props" {
										props = t.Name.Name
										return true
									}
								}
							}
							return false
						}
						if selectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {
							if searchComponent(selectorExpr) {
								break
							}
						} else {
							if indexExpr, ok := field.Type.(*ast.IndexExpr); ok {
								if selectorExpr, ok := indexExpr.X.(*ast.SelectorExpr); ok {
									if searchComponent(selectorExpr) {
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if component != "" || functionComponent != "" {
		for _, decl := range file.Decls {
			switch c := decl.(type) {
			case *ast.GenDecl:
				if c.Tok == token.IMPORT {
					hasJs := false
					hasChunks := false
					hasCopier := false
					for _, spec := range c.Specs {
						if !isWatch || spec.(*ast.ImportSpec).Path.Value == "\"github.com/gopherjs/gopherjs/chunks\"" {
							hasChunks = true
						}
						if spec.(*ast.ImportSpec).Path.Value == "\"github.com/gopherjs/gopherjs/js\"" {
							hasJs = true
						}
						if (!isWatch && component != "") || spec.(*ast.ImportSpec).Path.Value == "\"github.com/jinzhu/copier\"" {
							hasCopier = true
						}
						if hasJs && hasChunks && hasCopier {
							break
						}
					}
					if !hasChunks {
						c.Specs = append(c.Specs, &ast.ImportSpec{
							Path: &ast.BasicLit{Kind: token.STRING, Value: "\"github.com/gopherjs/gopherjs/chunks\""},
						})
					}
					if !hasJs {
						c.Specs = append(c.Specs, &ast.ImportSpec{
							Path: &ast.BasicLit{Kind: token.STRING, Value: "\"github.com/gopherjs/gopherjs/js\""},
						})
					}
					if !hasCopier {
						c.Specs = append(c.Specs, &ast.ImportSpec{
							Path: &ast.BasicLit{Kind: token.STRING, Value: "\"github.com/jinzhu/copier\""},
						})
					}
				}
			}
		}
		if functionComponent != "" {
			file.Decls = append(file.Decls, &ast.FuncDecl{
				Recv: &ast.FieldList{
					List: []*ast.Field{{
						Names: []*ast.Ident{{Name: "a"}},
						Type:  &ast.Ident{Name: functionComponent},
					},
					},
				},
				Name: &ast.Ident{Name: "HackRender"},
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{{
							Names: []*ast.Ident{{Name: "props"}},
							Type: &ast.StarExpr{
								X: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "js"},
									Sel: &ast.Ident{Name: "Object"},
								},
							},
						}},
					},
					Results: &ast.FieldList{
						List: []*ast.Field{{Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "react"},
							Sel: &ast.Ident{Name: "Element"},
						}}},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: "newProps"}},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.UnaryExpr{
							Op: token.AND,
							X:  &ast.CompositeLit{Type: &ast.Ident{Name: "FunProps"}},
						}},
					}, &ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "copier"},
								Sel: &ast.Ident{Name: "Copy"},
							},
							Args: []ast.Expr{
								&ast.Ident{Name: "newProps"},
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "react"},
										Sel: &ast.Ident{Name: "UnwrapValue"},
									},
									Args: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "props"},
												Sel: &ast.Ident{Name: "Get"},
											},
											Args: []ast.Expr{&ast.BasicLit{
												Kind:  token.STRING,
												Value: "\"_props\"",
											}},
										},
									},
								},
							},
						},
					}, &ast.ReturnStmt{
						Results: []ast.Expr{&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "a"},
								Sel: &ast.Ident{Name: "Default"},
							},
							Args: []ast.Expr{
								&ast.Ident{Name: "newProps"},
							},
						}},
					}},
				},
			})
		}

		if isWatch {
			if component != "" {
				getProps := &ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{{
							Names: []*ast.Ident{{Name: "a"}},
							Type:  &ast.Ident{Name: component},
						}},
					},
					Name: &ast.Ident{Name: "Props"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{{Type: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "react"},
								Sel: &ast.Ident{Name: "Props"},
							}}},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "props"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.UnaryExpr{
									Op: token.AND,
									X: &ast.CompositeLit{
										Type: &ast.Ident{
											Name: props,
										},
									},
								}},
							},
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "copier"},
										Sel: &ast.Ident{Name: "Copy"},
									},
									Args: []ast.Expr{
										&ast.Ident{Name: "props"},
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X: &ast.SelectorExpr{
													X:   &ast.Ident{Name: "a"},
													Sel: &ast.Ident{Name: "ComponentDef"},
												},
												Sel: &ast.Ident{Name: "Props"},
											},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{Name: "props"},
								},
							},
						},
					},
				}
				file.Decls = append(file.Decls, getProps)
			}

			var cmp string
			if component != "" {
				cmp = component
			} else {
				cmp = functionComponent
			}

			module := &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{{Name: "_"}},
						Values: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.ParenExpr{
									X: &ast.FuncLit{
										Type: &ast.FuncType{
											Params:  &ast.FieldList{},
											Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "bool"}}}},
										},
										Body: &ast.BlockStmt{
											List: []ast.Stmt{
												&ast.AssignStmt{
													Lhs: []ast.Expr{&ast.IndexExpr{
														X: &ast.SelectorExpr{
															X:   &ast.Ident{Name: "chunks"},
															Sel: &ast.Ident{Name: "GoChunks"},
														},
														Index: &ast.BasicLit{
															Kind:  token.STRING,
															Value: importPathValue,
														},
													}},
													Tok: token.ASSIGN,
													Rhs: []ast.Expr{&ast.FuncLit{
														Type: &ast.FuncType{
															Params: &ast.FieldList{},
															Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.InterfaceType{
																Methods: &ast.FieldList{},
															}}}},
														},
														Body: &ast.BlockStmt{
															List: []ast.Stmt{
																&ast.ReturnStmt{
																	Results: []ast.Expr{
																		&ast.UnaryExpr{
																			Op: token.AND,
																			X: &ast.CompositeLit{
																				Type: &ast.Ident{Name: cmp},
																			},
																		},
																	},
																},
															},
														},
													}},
												},
												&ast.ReturnStmt{
													Results: []ast.Expr{&ast.Ident{Name: "true"}},
												},
											},
										},
									},
								},
								Args: []ast.Expr{},
							},
						},
					},
				},
			}
			file.Decls = append(file.Decls, module)
		}
	}
}
