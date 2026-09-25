package conv

// PtrValue 返回指针指向的值，指针为 nil 时返回类型零值。
func PtrValue[T any](v *T) (t T) {
	if v != nil {
		t = *v
	}
	return
}
