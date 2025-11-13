package bitcask

import (
	"math/rand"
	"sort"
	"time"
)

// ====================================================================
// 常量定义
// ====================================================================

const (
	P             float64 = 0.5 // 升级概率
	MAX_LEVEL     int     = 16  // SkipList 最大层数
	MAX_ARRAY_LEN int     = 128 // 每个节点内部数组的最大长度
)

// ====================================================================
// Node 结构体
// ====================================================================

// Node 结构体：包含一个有序数组和多层指针
type Node struct {
	// 存储有序数据
	array Entries
	// forward[i] 表示当前节点在 i 层的下一个节点
	forward []*Node
}

// NewNode 创建一个新的 Node
func NewNode(val *Entry, level int) *Node {
	n := &Node{
		array:   make(Entries, 0, MAX_ARRAY_LEN),
		forward: make([]*Node, level+1),
	}
	n.array = append(n.array, val)
	return n
}

// First 返回数组的第一个元素
func (n *Node) First() *Entry {
	if len(n.array) == 0 {
		return nil // 逻辑上不应发生，但作为哨兵使用
	}
	return n.array[0]
}

// Last 返回数组的最后一个元素
func (n *Node) Last() *Entry {
	if len(n.array) == 0 {
		return nil // 逻辑上不应发生
	}
	return n.array[len(n.array)-1]
}

// Pop 从数组末尾弹出一个元素
func (n *Node) Pop() *Entry {
	if len(n.array) == 0 {
		return nil
	}
	last := n.array[len(n.array)-1]
	n.array = n.array[:len(n.array)-1]
	return last
}

// InsertIntoArray 在内部数组中插入元素，保持有序
func (n *Node) InsertIntoArray(key *Entry) {
	// 使用 sort.Search 查找插入位置
	pos := sort.Search(len(n.array), func(i int) bool { return n.array[i].GreaterEq(key) })

	// 插入
	n.array = append(n.array, nil)
	copy(n.array[pos+1:], n.array[pos:])
	n.array[pos] = key
}

// SearchInArray 在内部数组中查找元素
func (n *Node) SearchInArray(key *Entry) *Entry {
	// 使用 sort.Search 查找位置
	pos := sort.Search(len(n.array), func(i int) bool { return n.array[i].GreaterEq(key) })
	if pos < len(n.array) && n.array[pos].Equal(key) {
		return n.array[pos]
	}
	return nil // 未找到
}

// // IsFull 检查数组是否已满
func (n *Node) IsFull() bool {
	return len(n.array) >= MAX_ARRAY_LEN
}

// // IsEmpty 检查数组是否为空
func (n *Node) IsEmpty() bool {
	return len(n.array) == 0
}

// // String 实现 fmt.Stringer 接口用于显示
// func (n *Node) String() string {
// 	return fmt.Sprintf("%v", n.array)
// }

// // DeleteFromArray 从内部数组中删除指定的 key
func (n *Node) DeleteFromArray(key *Entry) bool {
	// 使用 sort.Search 查找位置
	pos := sort.Search(len(n.array), func(i int) bool { return n.array[i].GreaterEq(key) })

	// 检查是否找到
	if pos < len(n.array) && n.array[pos] == key {
		// 删除元素：将 pos 之后的元素向前移动一位
		copy(n.array[pos:], n.array[pos+1:])
		// 截断切片
		n.array = n.array[:len(n.array)-1]
		return true
	}
	return false // 未找到
}

// ====================================================================
// SkipListArr 结构体
// ====================================================================
type SkipListIterator struct {
	currentNode *Node
	index       int
	end         *Entry
}

// SkipListArr 结构体：数组跳表
type SkipListArr struct {
	level  int        // 当前最高层数
	header *Node      // 哨兵节点
	rand   *rand.Rand // 随机数生成器
}

// NewSkipListArr 创建一个新的数组跳表
func NewSkipListArr() *SkipListArr {
	// 使用一个固定但合理的种子，以便在基准测试中获得稳定的伪随机结果
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// 哨兵节点，包含 MIN_INT，层数为 MAX_LEVEL
	header := NewNode(nil, MAX_LEVEL)
	// 初始化 header 节点的数组，包含负无穷
	header.array = []*Entry{nil}

	return &SkipListArr{
		level:  0,
		header: header,
		rand:   r,
	}
}

// randomLevel 随机生成层数
func (s *SkipListArr) randomLevel() int {
	lvl := 0
	for s.rand.Float64() < P && lvl < MAX_LEVEL {
		lvl++
	}
	return lvl
}

// Insert 插入操作
func (s *SkipListArr) Insert(key *Entry) {
	update := make([]*Node, MAX_LEVEL+1)
	current := s.header

	// 1. 查找插入位置
	for i := s.level; i >= 0; i-- {
		// 查找在当前层 i 中，第一个节点的 First() 大于 key 的位置
		for current.forward[i] != nil && current.forward[i].First().LessEq(key) {
			current = current.forward[i]
		}
		update[i] = current
	}

	// 此时 current 是底层链表中 First() <= key 的最后一个节点

	// 2. 节点内部处理
	if current != s.header {
		// 优化：如果当前节点未满，直接插入到内部数组
		if !current.IsFull() {
			current.InsertIntoArray(key)
			return
		}

		// 溢出替换处理 (key 应该插入到 current 内部，但 current 已满)
		// 确保 key 不大于 current 的最大值，否则应插入到下一个节点
		if key.Less(current.Last()) {
			// 弹出 current 的最大值，作为新的 key 待插入
			replace := current.Pop()
			current.InsertIntoArray(key)
			key = replace // 继续用这个溢出的值执行后续的节点插入逻辑
		} else {
			// 如果 key >= current.Last()，说明 key 应该插入到 current 之后的节点
			// 但我们仍然需要检查 key 是否属于下一个节点，如果没有下一个节点，
			// 此时 current 节点已满，key 只能触发新节点创建。
		}
	}
	// 3. 检查下一个节点是否能容纳 key (如果 current 满了)
	if current.forward[0] != nil && !current.forward[0].IsFull() {
		current.forward[0].InsertIntoArray(key)
		return
	}

	// 4. 创建新节点并链入跳表
	newLevel := s.randomLevel()

	if newLevel > s.level {
		for i := s.level + 1; i <= newLevel; i++ {
			update[i] = s.header
		}
		s.level = newLevel
	}

	newNode := NewNode(key, newLevel)

	for i := 0; i <= newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}
}

// Search 查找操作
func (s *SkipListArr) Search(key *Entry) *Entry {
	current := s.header

	// 从最高层开始查找
	for i := s.level; i >= 0; i-- {
		// 查找在当前层 i 中，第一个节点的 First() 大于 key 的位置
		for current.forward[i] != nil && current.forward[i].First().LessEq(key) {
			current = current.forward[i]
		}
	}
	// 此时 current 是底层链表中 First() <= key 的最后一个节点

	// 在 current 节点内部的数组中查找
	if current != s.header {
		return current.SearchInArray(key)
	}
	return nil // 未找到
}

// RangeIterator 返回位于 [start, end] 的迭代器
func (s *SkipListArr) RangeIterator(start, end *Entry) *SkipListIterator {
	current := s.header

	// 1. 从最高层找到 >= start 的起始节点
	for i := s.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].Last().Less(start) {
			current = current.forward[i]
		}
	}

	current = current.forward[0] // 底层节点可能包含 start

	// 2. 定位节点内部数组中 >= start 的索引
	idx := 0
	if current != nil {
		idx = sort.Search(len(current.array), func(i int) bool {
			return current.array[i].GreaterEq(start)
		})
	}

	return &SkipListIterator{
		currentNode: current,
		index:       idx,
		end:         end,
	}
}

// Next 返回下一个 Entry，如果没有更多元素返回 nil
func (it *SkipListIterator) Next() *Entry {
	for it.currentNode != nil {
		if it.index >= len(it.currentNode.array) {
			// 当前节点遍历完，移动到下一个节点
			it.currentNode = it.currentNode.forward[0]
			it.index = 0
			continue
		}

		entry := it.currentNode.array[it.index]
		it.index++

		if entry == nil {
			continue
		}

		if entry.Greater(it.end) {
			// 超过范围，迭代结束
			it.currentNode = nil
			return nil
		}

		return entry
	}
	return nil
}

// Delete 从跳表中删除指定的 key
func (s *SkipListArr) Delete(key *Entry) bool {
	update := make([]*Node, MAX_LEVEL+1)
	current := s.header

	for i := s.level; i >= 0; i-- {
		// 查找在当前层 i 中，下一个节点 First() < key 的位置
		for current.forward[i] != nil && current.forward[i].First().Less(key) {
			current = current.forward[i]
		}
		update[i] = current // update[i] 是 key 所在节点的前驱节点
	}

	// current 此时是底层链表中 First() < key 的最后一个节点 (key 所在节点的前一个节点)
	current = current.forward[0] // current 现在指向可能包含 key 的节点

	if current == nil {
		return false // 没有找到节点
	}

	// 尝试在 current 节点内部删除 key
	res := current.DeleteFromArray(key)
	if !res {
		return false // 节点存在，但 key 不在内部数组中
	}

	// 如果删除成功且节点为空，则从跳表中移除该节点
	if current.IsEmpty() {
		// 遍历 level，更新指针，绕过 current 节点
		for i := 0; i <= s.level; i++ {
			if update[i].forward[i] != current {
				// 在较低层也找不到 current，说明 current 只存在于更高的层
				// 或者已经处理完毕，可以 break
				break
			}
			update[i].forward[i] = current.forward[i]
		}

		// (Go 语言不需要 delete)

		// 更新跳表的最高层数 level
		for s.level > 0 && s.header.forward[s.level] == nil {
			s.level--
		}
	}

	return true
}
