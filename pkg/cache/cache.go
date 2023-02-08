package cache

type Cache interface {
	Has(id string) bool
	Add(id string)
}
