package template

import (
	"fmt"
	html "html/template"
	"sort"
	text "text/template"
	"text/template/parse"
)

func ValidateTextTemplate(template *text.Template) error {
	tpls := template.Templates()
	sort.Slice(tpls, func(i, j int) bool {
		return tpls[i].Name() < tpls[j].Name()
	})

	for _, tpl := range tpls {
		if err := validateTree(tpl.Tree); err != nil {
			return err
		}
	}
	return nil
}

func ValidateHTMLTemplate(template *html.Template) error {
	tpls := template.Templates()
	sort.Slice(tpls, func(i, j int) bool {
		return tpls[i].Name() < tpls[j].Name()
	})

	for _, tpl := range tpls {
		if err := validateTree(tpl.Tree); err != nil {
			return err
		}
	}
	return nil
}

func validateTree(tree *parse.Tree) (err error) {
	validateFn := func(n parse.Node, depth int) (cont bool) {
		if depth > 4 {
			err = fmt.Errorf("%s: template nested too deep", formatLocation(tree, n))
		} else {
			switch n := n.(type) {
			case *parse.IfNode, *parse.ListNode, *parse.ActionNode, *parse.TextNode:
				break
			case *parse.PipeNode:
				if len(n.Decl) > 0 {
					err = fmt.Errorf("%s: declaration is forbidden", formatLocation(tree, n))
				} else if len(n.Cmds) > 1 {
					err = fmt.Errorf("%s: pipeline is forbidden", formatLocation(tree, n))
				}
			case *parse.CommandNode:
				for _, arg := range n.Args {
					if ident, ok := arg.(*parse.IdentifierNode); ok {
						if !checkIdentifier(ident.Ident) {
							err = fmt.Errorf("%s: forbidden identifier %s", formatLocation(tree, n), ident.Ident)
							break
						}
					}
				}
			default:
				err = fmt.Errorf("%s: forbidden construct %T", formatLocation(tree, n), n)
			}
		}

		return err == nil
	}

	traverseTree(tree, validateFn)
	return
}

var badIdentifiers = []string{
	"print",
	"printf",
	"println",
}

func checkIdentifier(id string) bool {
	for _, badID := range badIdentifiers {
		if id == badID {
			return false
		}
	}
	return true
}

func formatLocation(tree *parse.Tree, n parse.Node) string {
	location, _ := tree.ErrorContext(n)
	return location
}

func traverseTree(tree *parse.Tree, fn func(n parse.Node, depth int) (cont bool)) {
	var visit func(n parse.Node, depth int) (cont bool)
	visitBranch := func(n *parse.BranchNode, depth int) (cont bool) {
		if cont = visit(n.Pipe, depth); !cont {
			return
		}
		if cont = visit(n.List, depth); !cont {
			return
		}
		if n.ElseList != nil {
			if cont = visit(n.ElseList, depth); !cont {
				return false
			}
		}
		return
	}
	visit = func(n parse.Node, depth int) (cont bool) {
		cont = fn(n, depth)
		if !cont {
			return
		}

		switch n := n.(type) {
		case *parse.IfNode:
			cont = visitBranch(&n.BranchNode, depth)
		case *parse.RangeNode:
			cont = visitBranch(&n.BranchNode, depth)
		case *parse.WithNode:
			cont = visitBranch(&n.BranchNode, depth)
		case *parse.ListNode:
			for _, n := range n.Nodes {
				if cont = visit(n, depth+1); !cont {
					break
				}
			}
		case *parse.ActionNode:
			cont = visit(n.Pipe, depth)
		case *parse.PipeNode:
			for _, cmd := range n.Cmds {
				if cont = visit(cmd, depth); !cont {
					break
				}
			}
		case *parse.CommandNode:
			for _, arg := range n.Args {
				if pipe, ok := arg.(*parse.PipeNode); ok {
					if cont = visit(pipe, depth+1); !cont {
						break
					}
				}
			}
		case *parse.TemplateNode, *parse.TextNode:
			break
		default:
			panic("unknown node type")
		}
		return
	}
	visit(tree.Root, 0)
}
