/*
 Supplier and acceptor
*/

package strings_storage

type StringsStorage interface {
	Supplier
	Consumer
}

type Supplier interface {
	String() string
	Len() int
}

type Consumer interface {
	Accept(string)
}

type Storage struct {
	index   int
	strings []string
}

func NewStorage() *Storage {
	retVal := new(Storage)
	retVal.strings = make([]string, 0)
	return retVal
}

func (storage *Storage) String() string {
	if len((*storage).strings) == 0 || (*storage).index == len((*storage).strings) {
		// no more strings in the storage
		return ""
	}
	index := (*storage).index
	(*storage).index++
	return (*storage).strings[index]
}

// empty strings are discarded
func (storage *Storage) Accept(s string) {
	if len(s) > 0 {
		(*storage).strings = append((*storage).strings, s)
	}
}

func (storage *Storage) Len() int {
	return len((*storage).strings)
}

func (storage *Storage) ResetPos() {
	(*storage).index = 0
}

func (storage *Storage) Empty() {
	(*storage).index = 0
	(*storage).strings = (*storage).strings[:0]
}

func (storage *Storage) PeekPos() int {
	return (*storage).index
}

func (storage *Storage) ToArray() []string {
	retVal := make([]string,0)
	for i := range storage.strings {
		retVal = append(retVal, storage.strings[i])
	}
	return retVal
}
