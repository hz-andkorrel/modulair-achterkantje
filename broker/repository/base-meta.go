package repository

// The base for the meta repository takes care of getting and setting information in the meta repository.
// Gettings returns the value of the key as string, setting returns the name of the key set.
// On error, both return an empty string.
type BaseMetaRepository interface {
	Get(key string) string
	Set(key string, value string) string
}
