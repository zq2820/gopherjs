package build

import (
	"fmt"
	"go/ast"
	"go/token"
)

func GenerateTemplate(file *ast.File, importPath string, isWatch bool) {
	importPathValue := fmt.Sprintf("\"%s\"", importPath)

	var component *ast.Ident
	var functionComponent *ast.Ident
	var props ast.Expr
	for _, decl := range file.Decls {
		if (component != nil || functionComponent != nil) && props != nil {
			break
		}

		switch c := decl.(type) {
		case *ast.GenDecl:
			if t, ok := c.Specs[0].(*ast.TypeSpec); ok {
				if fieldList, ok := t.Type.(*ast.StructType); ok {
					for _, field := range fieldList.Fields.List {
						if typeParams, ok := field.Type.(*ast.IndexListExpr); ok {
							if selectorExpr, ok := typeParams.X.(*ast.SelectorExpr); ok {
								if ident, ok := selectorExpr.X.(*ast.Ident); ok {
									if ident.Name == "react" {
										if selectorExpr.Sel.Name == "ComponentDef" {
											component = t.Name
										}
										props = typeParams.Indices[0]
										break
									}
								}
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			if c.Type.Results != nil && c.Recv == nil && len(c.Type.Results.List) == 1 {
				if selectorExpr, ok := c.Type.Results.List[0].Type.(*ast.SelectorExpr); ok {
					if x, ok := selectorExpr.X.(*ast.Ident); ok {
						if x.Name == "react" {
							if selectorExpr.Sel.Name == "Element" {
								functionComponent = c.Name
								if propsDef, ok := c.Type.Params.List[0].Type.(*ast.StarExpr); ok {
									props = propsDef.X
								} else {
									props = c.Type.Params.List[0].Type
								}
								break
							}
						}
					}
				}
			}
		}
	}

	if component != nil || functionComponent != nil {
		if props == nil {
			panic("Props is need in gox")
		}

		for _, decl := range file.Decls {
			switch c := decl.(type) {
			case *ast.GenDecl:
				if c.Tok == token.IMPORT {
					hasJs := !isWatch || functionComponent == nil
					hasChunks := !isWatch
					hasCopier := !isWatch
					hasReflect := !isWatch || functionComponent == nil
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
						if spec.(*ast.ImportSpec).Path.Value == "\"reflect\"" {
							hasReflect = true
						}
						if hasChunks && hasCopier {
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
					if !hasReflect {
						c.Specs = append(c.Specs, &ast.ImportSpec{
							Path: &ast.BasicLit{Kind: token.STRING, Value: "\"reflect\""},
						})
					}
				}
			}
		}

		if isWatch {
			if component != nil {
				getProps := &ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{{
							Names: []*ast.Ident{{Name: "a"}},
							Type:  component,
						}},
					},
					Name: &ast.Ident{Name: "Props"},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{{Type: &ast.StarExpr{
								X: props,
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
										Type: props,
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

			var cmp ast.Expr
			if component != nil {
				cmp = &ast.UnaryExpr{
					Op: token.AND,
					X: &ast.CompositeLit{
						Type: component,
					},
				}
			} else {
				cmp = &ast.FuncLit{
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{{
								Names: []*ast.Ident{{Name: "props"}},
								Type: &ast.InterfaceType{
									Methods: &ast.FieldList{},
								},
							}, {
								Names: []*ast.Ident{{Name: "children"}},
								Type: &ast.Ellipsis{
									Ellipsis: 1,
									Elt: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "react"},
										Sel: &ast.Ident{Name: "Element"},
									},
								}},
							},
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
								X:  &ast.CompositeLit{Type: props},
							}},
						}, &ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "copier"},
									Sel: &ast.Ident{Name: "Copy"},
								},
								Args: []ast.Expr{
									&ast.Ident{Name: "newProps"},
									&ast.Ident{Name: "props"},
								},
							},
						}, &ast.ReturnStmt{
							Results: []ast.Expr{&ast.CallExpr{
								Fun: functionComponent,
								Args: []ast.Expr{
									&ast.Ident{Name: "newProps"},
									&ast.Ident{Name: "children"},
								},
								Ellipsis: 1,
							}},
						}},
					},
				}

				cmp = &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "js"},
							Sel: &ast.Ident{Name: "Global"},
						},
						Sel: &ast.Ident{Name: "Call"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"$makeFunc\"",
						},
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "js"},
								Sel: &ast.Ident{Name: "InternalObject"},
							},
							Args: []ast.Expr{
								&ast.FuncLit{
									Type: &ast.FuncType{
										Params: &ast.FieldList{
											List: []*ast.Field{
												{
													Names: []*ast.Ident{{Name: "this"}},
													Type: &ast.StarExpr{
														X: &ast.SelectorExpr{
															X:   &ast.Ident{Name: "js"},
															Sel: &ast.Ident{Name: "Object"},
														},
													},
												},
												{
													Names: []*ast.Ident{{Name: "arguments"}},
													Type: &ast.ArrayType{Elt: &ast.StarExpr{
														X: &ast.SelectorExpr{
															X:   &ast.Ident{Name: "js"},
															Sel: &ast.Ident{Name: "Object"},
														},
													}},
												},
											},
										},
										Results: &ast.FieldList{
											List: []*ast.Field{{Type: &ast.InterfaceType{
												Methods: &ast.FieldList{},
											}}},
										},
									},
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											&ast.AssignStmt{
												Lhs: []ast.Expr{
													&ast.Ident{Name: "props"},
												},
												Tok: token.DEFINE,
												Rhs: []ast.Expr{
													&ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X:   &ast.Ident{Name: "js"},
															Sel: &ast.Ident{Name: "UnwrapValue"},
														},
														Args: []ast.Expr{
															&ast.CallExpr{
																Fun: &ast.SelectorExpr{
																	X: &ast.IndexExpr{
																		X: &ast.Ident{Name: "arguments"},
																		Index: &ast.BasicLit{
																			Kind:  token.INT,
																			Value: "0",
																		},
																	},
																	Sel: &ast.Ident{Name: "Get"},
																},
																Args: []ast.Expr{
																	&ast.BasicLit{
																		Kind:  token.STRING,
																		Value: "\"_props\"",
																	},
																},
															},
														},
													},
												},
											},
											&ast.AssignStmt{
												Lhs: []ast.Expr{&ast.Ident{Name: "unwrapChildren"}},
												Tok: token.DEFINE,
												Rhs: []ast.Expr{
													&ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X:   &ast.Ident{Name: "reflect"},
															Sel: &ast.Ident{Name: "ValueOf"},
														},
														Args: []ast.Expr{
															&ast.StarExpr{
																X: &ast.TypeAssertExpr{
																	X: &ast.CallExpr{
																		Fun: &ast.SelectorExpr{
																			X:   &ast.Ident{Name: "js"},
																			Sel: &ast.Ident{Name: "UnwrapValue"},
																		},
																		Args: []ast.Expr{
																			&ast.CallExpr{
																				Fun: &ast.SelectorExpr{
																					X: &ast.IndexExpr{
																						X: &ast.Ident{Name: "arguments"},
																						Index: &ast.BasicLit{
																							Kind:  token.INT,
																							Value: "0",
																						},
																					},
																					Sel: &ast.Ident{Name: "Get"},
																				},
																				Args: []ast.Expr{
																					&ast.BasicLit{
																						Kind:  token.STRING,
																						Value: "\"_children\"",
																					},
																				},
																			},
																		},
																	},
																	Type: &ast.StarExpr{
																		X: &ast.ArrayType{
																			Elt: &ast.SelectorExpr{
																				X:   &ast.Ident{Name: "react"},
																				Sel: &ast.Ident{Name: "Element"},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
											&ast.AssignStmt{
												Lhs: []ast.Expr{&ast.Ident{Name: "unwrapArgs"}},
												Tok: token.DEFINE,
												Rhs: []ast.Expr{&ast.CallExpr{
													Fun: &ast.Ident{Name: "make"},
													Args: []ast.Expr{
														&ast.ArrayType{
															Elt: &ast.SelectorExpr{
																X:   &ast.Ident{Name: "reflect"},
																Sel: &ast.Ident{Name: "Value"},
															},
														},
														&ast.BasicLit{
															Kind:  token.INT,
															Value: "0",
														},
														&ast.BinaryExpr{
															X: &ast.CallExpr{
																Fun: &ast.SelectorExpr{
																	X:   &ast.Ident{Name: "unwrapChildren"},
																	Sel: &ast.Ident{Name: "Len"},
																},
															},
															Op: token.ADD,
															Y: &ast.BasicLit{
																Kind:  token.INT,
																Value: "1",
															},
														},
													},
												}},
											},
											&ast.AssignStmt{
												Lhs: []ast.Expr{&ast.Ident{Name: "newProps"}},
												Tok: token.DEFINE,
												Rhs: []ast.Expr{&ast.UnaryExpr{
													Op: token.AND,
													X:  &ast.CompositeLit{Type: props},
												}},
											},
											&ast.ExprStmt{
												X: &ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "copier"},
														Sel: &ast.Ident{Name: "Copy"},
													},
													Args: []ast.Expr{
														&ast.Ident{Name: "newProps"},
														&ast.Ident{Name: "props"},
													},
												},
											},
											&ast.AssignStmt{
												Lhs: []ast.Expr{&ast.Ident{Name: "unwrapArgs"}},
												Tok: token.ASSIGN,
												Rhs: []ast.Expr{&ast.CallExpr{
													Fun: &ast.Ident{Name: "append"},
													Args: []ast.Expr{
														&ast.Ident{Name: "unwrapArgs"},
														&ast.CallExpr{
															Fun: &ast.SelectorExpr{
																X:   &ast.Ident{Name: "reflect"},
																Sel: &ast.Ident{Name: "ValueOf"},
															},
															Args: []ast.Expr{
																&ast.Ident{Name: "newProps"},
															},
														},
													},
												}},
											},
											&ast.ForStmt{
												Init: &ast.AssignStmt{
													Lhs: []ast.Expr{&ast.Ident{Name: "i"}},
													Tok: token.DEFINE,
													Rhs: []ast.Expr{&ast.BasicLit{
														Kind:  token.INT,
														Value: "0",
													}},
												},
												Cond: &ast.BinaryExpr{
													X:  &ast.Ident{Name: "i"},
													Op: token.LSS,
													Y: &ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X:   &ast.Ident{Name: "unwrapChildren"},
															Sel: &ast.Ident{Name: "Len"},
														},
													},
												},
												Post: &ast.IncDecStmt{
													X:   &ast.Ident{Name: "i"},
													Tok: token.INC,
												},
												Body: &ast.BlockStmt{
													List: []ast.Stmt{
														&ast.AssignStmt{
															Lhs: []ast.Expr{&ast.Ident{Name: "unwrapArgs"}},
															Tok: token.ASSIGN,
															Rhs: []ast.Expr{&ast.CallExpr{
																Fun: &ast.Ident{Name: "append"},
																Args: []ast.Expr{
																	&ast.Ident{Name: "unwrapArgs"},
																	&ast.CallExpr{
																		Fun: &ast.SelectorExpr{
																			X:   &ast.Ident{Name: "unwrapChildren"},
																			Sel: &ast.Ident{Name: "Index"},
																		},
																		Args: []ast.Expr{
																			&ast.Ident{Name: "i"},
																		},
																	},
																},
															}},
														},
													},
												},
											},
											&ast.ReturnStmt{
												Results: []ast.Expr{
													&ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X: &ast.IndexExpr{
																X: &ast.CallExpr{
																	Fun: &ast.SelectorExpr{
																		X: &ast.CallExpr{
																			Fun: &ast.SelectorExpr{
																				X:   &ast.Ident{Name: "reflect"},
																				Sel: &ast.Ident{Name: "ValueOf"},
																			},
																			Args: []ast.Expr{functionComponent},
																		},
																		Sel: &ast.Ident{Name: "Call"},
																	},
																	Args: []ast.Expr{&ast.Ident{Name: "unwrapArgs"}},
																},
																Index: &ast.BasicLit{
																	Kind:  token.INT,
																	Value: "0",
																},
															},
															Sel: &ast.Ident{Name: "Interface"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("\"%s\"", functionComponent.Name),
						},
					},
				}
			}

			init := &ast.FuncDecl{
				Name: &ast.Ident{
					Name: "init",
				},
				Type: &ast.FuncType{
					Params: &ast.FieldList{},
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
							Rhs: []ast.Expr{cmp},
						},
					},
				},
			}

			_ = importPathValue
			_ = cmp
			file.Decls = append(file.Decls, init)
		}
	}
}
