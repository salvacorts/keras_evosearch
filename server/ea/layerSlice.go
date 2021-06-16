package ea

import (
	pb "server/protobuf/api"

	"github.com/MaxHalford/eaopt"
)

type Layers []*pb.Layer

func (l Layers) At(i int) interface{} {
	return l[i]
}

func (l Layers) Set(i int, v interface{}) {
	l[i] = v.(*pb.Layer)
}

func (l Layers) Len() int {
	return len(l)
}

func (l Layers) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l Layers) Slice(a, b int) eaopt.Slice {
	return l[a:b]
}

func (l Layers) Split(k int) (eaopt.Slice, eaopt.Slice) {
	return l[:k], l[k:]
}

func (l Layers) Append(q eaopt.Slice) eaopt.Slice {
	return append(l, q.(Layers)...)
}

func (l Layers) Replace(q eaopt.Slice) {
	copy(l, q.(Layers))
}

func (l Layers) Copy() eaopt.Slice {
	var clone = make(Layers, len(l))
	copy(clone, l)
	return clone
}
