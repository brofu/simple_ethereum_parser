package parser

type addressTransactionNode struct {
	addressTransaction
	previous *addressTransactionNode
	next     *addressTransactionNode
}

type addressTransactionLRU struct {
	capability int
	dataMap    map[string]*addressTransactionNode
	head       *addressTransactionNode
	tail       *addressTransactionNode
}

func newAddressTransactionLRU(capability int) *addressTransactionLRU {
	lru := &addressTransactionLRU{
		capability: capability,
		dataMap:    make(map[string]*addressTransactionNode, capability),
		head:       &addressTransactionNode{},
		tail:       &addressTransactionNode{},
	}
	lru.head.next = lru.tail
	lru.tail.previous = lru.head

	return lru
}

func (this *addressTransactionLRU) putAddress(data addressTransaction) {

	node, ok := this.dataMap[data.address]
	if ok {
		// already exist. Usaually, need to udpate the existing data,
		// but in our scenarios, we need keep the old data, since the new addresses would be added without transactions
		return
	}

	// if exceed the capbility, remove the tail
	if len(this.dataMap) >= this.capability {
		this.removeTail()
	}

	node = &addressTransactionNode{addressTransaction: data}
	// insert new node into data map
	this.dataMap[data.address] = node
	// insert new node into list
	node.next = this.head.next
	this.head.next.previous = node
	this.head.next = node
	node.previous = this.head
}

func (this *addressTransactionLRU) getAddress(addr string) *addressTransaction {

	node, ok := this.dataMap[addr]
	if !ok {
		return nil
	}

	if node.previous == this.head { // already the head element
		return &node.addressTransaction
	}

	// adjust the location of node
	node.previous.next = node.next
	node.next.previous = node.previous

	node.next = this.head.next
	this.head.next.previous = node

	node.previous = this.head
	this.head.next = node

	return &node.addressTransaction
}

func (this *addressTransactionLRU) size() int {
	return len(this.dataMap)
}

func (this *addressTransactionLRU) removeTail() {
	node := this.tail.previous

	if node == this.head { // this should NOT happen in real world
		return
	}

	delete(this.dataMap, node.addressTransaction.address) // remove from the map

	// remove from the list
	node.previous.next = node.next
	node.next.previous = node.previous
}

// getAddressIn would not affect the `frequency` of an node when it's accessed
func (this *addressTransactionLRU) getAddressIn(addr string) *addressTransaction {
	if node, ok := this.dataMap[addr]; ok {
		return &node.addressTransaction
	} else {
		return nil
	}
}

func (this *addressTransactionLRU) allAddresses() []string {
	addresses := make([]string, len(this.dataMap))
	for i, node := 0, this.head.next; i < len(addresses) && node != nil; i, node = i+1, node.next {
		addresses[i] = node.addressTransaction.address
	}
	return addresses
}
