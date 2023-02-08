package kubernetes

type Cache interface {
	AddItem(string, interface{})
	GetItem(string) (interface{}, bool)
	RemoveItem(string)
}
