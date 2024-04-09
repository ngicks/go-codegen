package clonehelper

func CopyPtr[V comparable](v *V) *V {
	if v == nil {
		return nil
	}
	vv := *v
	return &vv
}

type Cloner[V any] interface {
	Clone() V
}

func ClonePtr[V any, U Cloner[V]](v *U) *V {
	if v == nil {
		return nil
	}
	vv := (*v).Clone()
	return &vv
}
