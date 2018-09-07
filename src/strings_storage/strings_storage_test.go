package strings_storage

import (
	"log"
	"os"
	"strings"
	"testing"
)

var testArray = []string{
	"string 0",
	"string 1",
	"string 2",
	"string 3",
	"string 4",
	"string 5",
	"string 6",
	"string 7",
	"string 8",
	"string 9",
	"string 10",
	"string 11",
	"string 12",
	"string 13",
	"string 14",
	"string 15",
	"string 16",
	"string 17",
	"string 18",
}

var ext_storage *Storage

func TestMain(m *testing.M) {
	ext_storage = NewStorage()
	if ext_storage.Len() != 0 {
		log.Println("ext_storage.Len() == 0 failed")
	} else {
		log.Println("ext_storage.Len() == 0 passed")
	}
	os.Exit(m.Run())
}

func TestStorage_String(t *testing.T) {

	if strings.Compare(ext_storage.String(), "") != 0 {
		t.Error("reading from the empty storage error")
	} else {
		log.Println("reading from the empty storage passed")
	}

}

func TestNewStorage(t *testing.T) {
	const arrLen int = 100000
	var storageArray [arrLen]*Storage

	for i := 0; i < arrLen; i++ {
		storageArray[i] = NewStorage()
	}

	for i := range storageArray {
		for _, inString := range testArray {
			storageArray[i].Accept(inString)
		}
		if storageArray[i].Len() != len(testArray) {
			t.Error("storageArray[i].Len() != len(testArray)")
		} else {
			//			log.Println("PASSED: storageArray[i].Len() == len(testArray)")
		}
	}

	for j := range testArray {
		for i := range storageArray {
			if strings.Compare(testArray[j], storageArray[i].String()) != 0 {
				t.Error("testArray[j] not equal storageArray[i].String()")
			}
		}
	}
	// try to read beyong storage size
	for i := range storageArray {
		if strings.Compare("", storageArray[i].String()) != 0 {
			t.Error("read beyond storage size returned non-empty string!")
		}
	}
	// reset indexes
	for i := range storageArray {
		storageArray[i].ResetPos()
		}
	t.Log("read again after resetting positions")
	for j := range testArray {
		for i := range storageArray {
			if strings.Compare(testArray[j], storageArray[i].String()) != 0 {
				t.Error("testArray[j] not equal storageArray[i].String()")
			}
		}
	}

	// trying to empty
	t.Log("trying to empty all the storages")
	for i := range storageArray {
		storageArray[i].Empty()
	}
	// try to read from empty
	for i := range storageArray {
		if strings.Compare("", storageArray[i].String()) != 0 {
			t.Error("read beyond storage size returned non-empty string!")
		}
	}

	// try to store strings again
	t.Log("trying to store strings again")
	for i := range storageArray {
		for _, inString := range testArray {
			storageArray[i].Accept(inString)
		}
		if storageArray[i].Len() != len(testArray) {
			t.Error("storageArray[i].Len() != len(testArray)")
		} else {
			//			log.Println("PASSED: storageArray[i].Len() == len(testArray)")
		}
	}

	for j := range testArray {
		for i := range storageArray {
			if strings.Compare(testArray[j], storageArray[i].String()) != 0 {
				t.Error("testArray[j] not equal storageArray[i].String()")
			}
		}
	}
/*
	t.Log("for loop test")
	storageArray[0].ResetPos()
	s := storageArray[0].String()
	for len(s) > 0 {
		t.Log(s)
		s = storageArray[0].String()
	}
*/

	t.Log("ToArray() test")

	result := storageArray[0].ToArray()

	storageArray[0].ResetPos()

	result[0] = "Bla-Bla-Bla"

	if strings.Compare(storageArray[0].String(), result[0]) == 0 {
		t.Error("we need to copy deeper")
	}
}
