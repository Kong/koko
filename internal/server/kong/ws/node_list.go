package ws

import (
	"fmt"
	"sync"
)

type NodeList struct {
	nodes sync.Map
}

func (l *NodeList) Add(node *Node) error {
	remoteAddr := node.conn.RemoteAddr().String()
	_, loaded := l.nodes.LoadOrStore(remoteAddr, node)
	if loaded {
		return fmt.Errorf("node(ip: %v) already present", remoteAddr)
	}
	return nil
}

func (l *NodeList) Remove(node *Node) error {
	remoteAddr := node.conn.RemoteAddr().String()
	_, loaded := l.nodes.LoadAndDelete(remoteAddr)
	if !loaded {
		return fmt.Errorf("node(ip: %v) not found", remoteAddr)
	}
	return nil
}

func (l *NodeList) All() []*Node {
	var res []*Node
	l.nodes.Range(func(key, value interface{}) bool {
		node, ok := value.(*Node)
		if !ok {
			panic(fmt.Sprintf("expected type %T but got %T", Node{}, value))
		}
		res = append(res, node)
		return true
	})
	return res
}
