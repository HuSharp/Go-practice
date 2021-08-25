package gee

import (
	"fmt"
	"strings"
)

type node struct {
	pattern		string	// 待匹配的路由，例如 /p/lang   /p/:type
	part		string	// 路由中的一部分，例如 lang  :type
	children	[]*node	// 子节点，例如 [doc, tutorial, intro]
	isWild		bool	// 是否精确匹配，part 含有 : 或 * 时为true
}
/*
与普通的树不同，为了实现动态路由匹配，加上了 isWild 这个参数。
即当我们匹配 /p/go/doc/这个路由时，
第一层节点，p 精准匹配到了 p，第二层节点，go模糊匹配到:lang，
那么将会把lang这个参数赋值为go，继续下一层匹配。我们将匹配的逻辑，包装为一个辅助函数。
 */

func (n *node) String() string {
	return fmt.Sprintf("node{pattern=%s, part=%s, isWild=%t}", n.pattern, n.part, n.isWild)
}

// 表示第一个匹配成功的节点
func (n *node) matchChild(part string)	*node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 返回所有匹配成功的节点
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, node := range n.children {
		if node.part == part || node.isWild {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// 插入到 trie 中
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern	// 只有在最后匹配节点，才会将 pattern 设置为查询节点
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part: part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height + 1)

}
/*
/p/:lang/doc只有在第三层节点，即doc节点，pattern才会设置为/p/:lang/doc。
p和:lang节点的pattern属性皆为空。
因此，当匹配结束时，我们可以使用n.pattern == ""来判断路由规则是否匹配成功。
例如，/p/python虽能成功匹配到:lang，但:lang的pattern值为空，因此匹配失败。
 */
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)// 返回所有匹配节点

	for _, child := range children {
		// 继续递归查询
		result := child.search(parts, height + 1)
		if result != nil {
			return result
		}
	}
	return nil
}

func (n *node) travel(list *([]*node)) {
	if n.pattern != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}