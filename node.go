package nbmux

import (
	"fmt"
	"net/http"
	"regexp"
)

type Method uint8

const (
	GET     Method = 0
	HEAD    Method = 1
	POST    Method = 1 << 1
	PUT     Method = 1 << 2
	DELETE  Method = 1 << 3
	CONNECT Method = 1 << 4
	OPTIONS Method = 1 << 5
	TRACE   Method = 1 << 6
	ALL     Method = (1 << 8) - 1
)

type nbNode struct {
	parent       *nbNode
	childrenList []*nbNode
	childrenMap  map[string]map[Method]*nbNode
	method       Method
	handler      http.Handler
	pattern      *regexp.Regexp
}

func newRoot(notFoundHandler http.Handler) *nbNode {
	return &nbNode{
		parent:       nil,
		childrenList: make([]*nbNode, 0, 64),
		childrenMap:  make(map[string]map[Method]*nbNode),
		method:       ALL,
		handler:      notFoundHandler,
		pattern:      regexp.MustCompile(`.*`),
	}
}

func (node *nbNode) addChildren(exp []string, method Method, handler http.Handler) error {
	curExp, remExp := exp[0], exp[1:]
	pattern, err := regexp.Compile(curExp)
	if err != nil {
		return err
	}
	subMap, expExists := node.childrenMap[curExp]
	if expExists {
		existedNode, nodeExists := subMap[method]
		if nodeExists {
			if len(remExp) == 0 {
				if existedNode.handler != nil {
					return fmt.Errorf("%s regexp has already exists", curExp)
				} else {
					existedNode.handler = handler
				}
			} else {
				return existedNode.addChildren(remExp, method, handler)
			}
		} else {
			if len(remExp) == 0 {
				childNode := &nbNode{
					parent:       node,
					childrenList: make([]*nbNode, 0, 64),
					childrenMap:  make(map[string]map[Method]*nbNode),
					method:       method,
					handler:      handler,
					pattern:      pattern,
				}
				node.childrenList = append(node.childrenList, childNode)
				subMap[method] = childNode
			} else {
				childNode := &nbNode{
					parent:       node,
					childrenList: make([]*nbNode, 0, 64),
					childrenMap:  make(map[string]map[Method]*nbNode),
					method:       method,
					pattern:      pattern,
				}
				node.childrenList = append(node.childrenList, childNode)
				subMap[method] = childNode
				return childNode.addChildren(remExp, method, handler)
			}
		}
	} else {
		if len(remExp) == 0 {
			childNode := &nbNode{
				parent:       node,
				childrenList: make([]*nbNode, 0, 64),
				childrenMap:  make(map[string]map[Method]*nbNode),
				method:       method,
				handler:      handler,
				pattern:      pattern,
			}
			node.childrenList = append(node.childrenList, childNode)
			node.childrenMap[curExp] = map[Method]*nbNode{method: childNode}
		} else {
			childNode := &nbNode{
				parent:       node,
				childrenList: make([]*nbNode, 0, 64),
				childrenMap:  make(map[string]map[Method]*nbNode),
				method:       method,
				pattern:      pattern,
			}
			node.childrenList = append(node.childrenList, childNode)
			node.childrenMap[curExp] = map[Method]*nbNode{method: childNode}
			return childNode.addChildren(remExp, method, handler)
		}
	}
	return nil
}

func (node *nbNode) match(path string, method Method) bool {
	return node.pattern.MatchString(path) && node.method&method > 0
}

type result struct {
	depth   int
	handler http.Handler
}

func (node *nbNode) search(pathList []string, method Method) http.Handler {
	curPath, remPath := pathList[0], pathList[1:]
	for _, childNode := range node.childrenList {
		if childNode.match(curPath, method) {
			if len(remPath) == 0 {
				return childNode.handler
			} else {
				return childNode.search(remPath, method)
			}
		}
	}
	return nil
}
