package verkle

import lru "github.com/hashicorp/golang-lru"

const cacheSize = 500_000

type VisitedMap lru.Cache

func NewVisitedMap() *VisitedMap {
	c, err := lru.New(cacheSize)
	if err != nil {
		panic(err)
	}
	return (*VisitedMap)(c)
}

func (v *VisitedMap) Visited(singularity interface{}) bool {
	return (*lru.Cache)(v).Contains(singularity)
}

func (v *VisitedMap) Visit(singularity interface{}) {
	(*lru.Cache)(v).Add(singularity, true)
}
