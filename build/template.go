package build

import (
	"fmt"
	"go/ast"
	"go/token"
)

func GenerateTemplate(file *ast.File, importPath string, isWatch bool) {
	component := ""
	props := ""
	for _, decl := range file.Decls {
		if component != "" && props != "" {
			break
		}

		switch c := decl.(type) {
		case *ast.GenDecl:
			isComponent := false
			isProps := false
			if t, ok := c.Specs[0].(*ast.TypeSpec); ok {
				if fieldList, ok := t.Type.(*ast.StructType); ok {
					for _, field := range fieldList.Fields.List {
						if selectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {
							if ident, ok := selectorExpr.X.(*ast.Ident); ok {
								if ident.Name == "react" {
									if selectorExpr.Sel.Name == "ComponentDef" {
										isComponent = true
										break
									}
									if selectorExpr.Sel.Name == "Props" {
										isProps = true
										break
									}
								}
							}
						}
					}
				}
				if isComponent {
					component = t.Name.Name
				}
				if isProps {
					props = t.Name.Name
				}
			}
		}
	}

	if component != "" {
		buildComponentElem := "build" + component + "Elem"
		for _, decl := range file.Decls {
			switch c := decl.(type) {
			case *ast.FuncDecl:
				if c.Name.Name[0] >= 'A' && c.Name.Name[0] <= 'Z' {
					if c.Recv == nil {
						arg := "nil"
						if len(c.Type.Params.List) > 0 {
							arg = c.Type.Params.List[0].Names[0].Name
						}
						args := []ast.Expr{
							&ast.Ident{Name: arg},
						}
						if len(c.Type.Params.List) > 1 {
							args = append(args, &ast.Ident{Name: c.Type.Params.List[1].Names[0].Name})
						}

						var callExpr string
						if !isWatch {
							callExpr = buildComponentElem
						} else {
							callExpr = "build_go_react_hotElem"
						}
						c.Body = &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.CallExpr{
											Fun:      &ast.Ident{Name: callExpr},
											Args:     args,
											Ellipsis: 1,
										},
									},
								},
							},
						}
					}
				}
			case *ast.GenDecl:
				if isWatch && c.Tok == token.IMPORT {
					hasJs := false
					hasChunks := false
					hasCopier := false
					for _, spec := range c.Specs {
						if spec.(*ast.ImportSpec).Path.Value == "\"github.com/gopherjs/gopherjs/chunks\"" {
							hasChunks = true
						}
						if spec.(*ast.ImportSpec).Path.Value == "\"github.com/gopherjs/gopherjs/js\"" {
							hasJs = true
						}
						if spec.(*ast.ImportSpec).Path.Value == "\"github.com/jinzhu/copier\"" {
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
		initTemplate(file, component)

		if isWatch {
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

			importPathValue := fmt.Sprintf("\"%s\"", importPath)
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
													Rhs: []ast.Expr{&ast.Ident{Name: buildComponentElem}},
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

			/** 热更新包裹组件 */
			hotTypeSpec := &ast.TypeSpec{
				Name: &ast.Ident{Name: "_go_react_hot"},
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{{Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "react"},
							Sel: &ast.Ident{Name: "ComponentDef"},
						}}},
					},
				},
			}

			hotComponent := ast.NewObj(ast.Typ, "_go_react_hot")
			hotComponent.Decl = hotTypeSpec
			file.Scope.Objects["_go_react_hot"] = hotComponent

			file.Decls = append(file.Decls, &ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					hotTypeSpec,
				},
			})
			file.Decls = append(file.Decls, &ast.FuncDecl{
				Recv: &ast.FieldList{
					List: []*ast.Field{{
						Names: []*ast.Ident{{Name: "a"}},
						Type:  &ast.Ident{Name: "_go_react_hot"},
					}},
				},
				Name: &ast.Ident{Name: "Render"},
				Type: &ast.FuncType{
					Params: &ast.FieldList{},
					Results: &ast.FieldList{
						List: []*ast.Field{{Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "react"},
							Sel: &ast.Ident{Name: "Element"},
						}}},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.TypeAssertExpr{
										X: &ast.IndexExpr{
											X: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "chunks"},
												Sel: &ast.Ident{Name: "GoChunks"},
											},
											Index: &ast.BasicLit{
												Kind:  token.STRING,
												Value: importPathValue,
											},
										},
										Type: &ast.FuncType{
											Params: &ast.FieldList{
												List: []*ast.Field{{
													Names: []*ast.Ident{{Name: "props"}},
													Type: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "react"},
														Sel: &ast.Ident{Name: "Props"},
													},
												}, {
													Names: []*ast.Ident{{Name: "children"}},
													Type: &ast.Ellipsis{
														Elt: &ast.SelectorExpr{
															X:   &ast.Ident{Name: "react"},
															Sel: &ast.Ident{Name: "Element"},
														},
														Ellipsis: 1,
													},
												}},
											},
											Results: &ast.FieldList{
												List: []*ast.Field{{
													Type: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "react"},
														Sel: &ast.Ident{Name: "Element"},
													},
												}},
											},
										},
									},
									Args: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "a"},
												Sel: &ast.Ident{Name: "Props"},
											},
											Args: []ast.Expr{},
										},
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "a"},
												Sel: &ast.Ident{Name: "Children"},
											},
											Args: []ast.Expr{},
										},
									},
									Ellipsis: 1,
								},
							},
						},
					},
				},
			})
			file.Decls = append(file.Decls, &ast.FuncDecl{
				Recv: &ast.FieldList{
					List: []*ast.Field{{
						Names: []*ast.Ident{{Name: "a"}},
						Type:  &ast.Ident{Name: "_go_react_hot"},
					}},
				},
				Name: &ast.Ident{Name: "ComponentDidMount"},
				Type: &ast.FuncType{Params: &ast.FieldList{}},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: "dependencies"}},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "js"},
											Sel: &ast.Ident{Name: "Global"},
										},
										Sel: &ast.Ident{Name: "Get"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: "\"window\"",
										},
									},
								},
								Sel: &ast.Ident{Name: "Get"},
							},
							Args: []ast.Expr{&ast.BasicLit{
								Kind:  token.STRING,
								Value: "\"dependencies\"",
							}},
						}},
					},
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "dependencies"},
									Sel: &ast.Ident{Name: "Get"},
								},
								Args: []ast.Expr{&ast.BasicLit{
									Kind:  token.STRING,
									Value: importPathValue,
								}},
							},
							Op: token.EQL,
							Y: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "js"},
								Sel: &ast.Ident{Name: "Undefined"},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "dependencies"},
											Sel: &ast.Ident{Name: "Set"},
										},
										Args: []ast.Expr{&ast.BasicLit{
											Kind:  token.STRING,
											Value: importPathValue,
										}, &ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X: &ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X: &ast.SelectorExpr{
															X:   &ast.Ident{Name: "js"},
															Sel: &ast.Ident{Name: "Global"},
														},
														Sel: &ast.Ident{Name: "Get"},
													},
													Args: []ast.Expr{&ast.BasicLit{
														Kind:  token.STRING,
														Value: "\"Array\"",
													}},
												},
												Sel: &ast.Ident{Name: "New"},
											},
										}},
									},
								},
							},
						},
					},
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "dependencies"},
										Sel: &ast.Ident{Name: "Get"},
									},
									Args: []ast.Expr{&ast.BasicLit{
										Kind:  token.STRING,
										Value: importPathValue,
									}},
								},
								Sel: &ast.Ident{Name: "Call"},
							},
							Args: []ast.Expr{&ast.BasicLit{
								Kind:  token.STRING,
								Value: "\"push\"",
							}, &ast.Ident{
								Name: "a",
							}},
						},
					},
				}},
			})
			initTemplate(file, "_go_react_hot")
		}
	}
}

func initTemplate(file *ast.File, component string) {
	buildComponent := "build" + component
	buildComponentElem := "build" + component + "Elem"

	fun := ast.NewObj(ast.Fun, buildComponent)
	fun.Decl = &ast.FuncDecl{
		Name: &ast.Ident{Name: buildComponent},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
	if decl, ok := fun.Decl.(*ast.FuncDecl); ok {
		decl.Type.Params.List = append(decl.Type.Params.List, &ast.Field{
			Names: []*ast.Ident{{Name: "cd"}},
			Type: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "react"},
				Sel: &ast.Ident{Name: "ComponentDef"},
			},
		})
		decl.Type.Results.List = append(decl.Type.Results.List, &ast.Field{
			Type: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "react"},
				Sel: &ast.Ident{Name: "Component"},
			},
		})
		decl.Body.List = append(decl.Body.List, &ast.ReturnStmt{
			Results: []ast.Expr{&ast.CompositeLit{
				Type: &ast.Ident{Name: component},
				Elts: []ast.Expr{
					&ast.KeyValueExpr{
						Key:   &ast.Ident{Name: "ComponentDef"},
						Value: &ast.Ident{Name: "cd"},
					},
				},
			}},
		})
		file.Decls = append(file.Decls, decl)
	}
	file.Scope.Objects[buildComponent] = fun

	funElem := ast.NewObj(ast.Fun, buildComponentElem)
	funElem.Decl = &ast.FuncDecl{
		Name: &ast.Ident{Name: buildComponentElem},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "props"}},
						Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "react"},
							Sel: &ast.Ident{Name: "Props"},
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
	if decl, ok := funElem.Decl.(*ast.FuncDecl); ok {
		decl.Type.Params.List = append(decl.Type.Params.List, &ast.Field{
			Names: []*ast.Ident{{Name: "children"}},
			Type: &ast.Ellipsis{
				Elt: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "react"},
					Sel: &ast.Ident{Name: "Element"},
				},
			},
		})
		decl.Type.Results.List = append(decl.Type.Results.List, &ast.Field{
			Type: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "react"},
				Sel: &ast.Ident{Name: "Element"},
			},
		})
		decl.Body.List = append(decl.Body.List, &ast.ReturnStmt{
			Results: []ast.Expr{&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "react"},
					Sel: &ast.Ident{Name: "CreateElement"},
				},
				Args: []ast.Expr{
					&ast.Ident{Name: buildComponent},
					&ast.Ident{Name: "props"},
					&ast.Ident{Name: "children"},
				},
				Ellipsis: 1,
			}},
		})
		file.Decls = append(file.Decls, decl)
	}
	file.Scope.Objects[buildComponentElem] = funElem

	renderFun := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "a"}},
				Type:  &ast.Ident{Name: component},
			}},
		},
		Name: &ast.Ident{Name: "RendersElement"},
		Type: &ast.FuncType{
			Results: &ast.FieldList{
				List: []*ast.Field{{
					Type: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "react"},
						Sel: &ast.Ident{Name: "Element"},
					},
				}},
			},
			Params: &ast.FieldList{
				List: []*ast.Field{},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
	renderFun.Body.List = append(renderFun.Body.List, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "a"},
					Sel: &ast.Ident{Name: "Render"},
				},
			},
		},
	})
	file.Decls = append(file.Decls, renderFun)
}
