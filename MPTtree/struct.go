package MPTtree

type fullNode struct {
	Children [17]node
	flags nodeFlag
}

type nodeFlag struct {
	hash hashNode
	gen uint16
	dirty bool
}

type shortNode struct {
	Key []byte
	Val node
	flags nodeFlag
}
