package git

/*
#include <git2.h>
#include <git2/errors.h>

extern int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr);
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type Tree struct {
	gitObject
}

type TreeEntry struct {
	Name string
	Id  *Oid
	Type ObjectType
}

func newTreeEntry(entry *C.git_tree_entry) *TreeEntry {
	return &TreeEntry{
		C.GoString(C.git_tree_entry_name(entry)),
		newOidFromC(C.git_tree_entry_id(entry)),
		ObjectType(C.git_tree_entry_type(entry)),
	}
}

func (t Tree) EntryByName(filename string) *TreeEntry {
	cname := C.CString(filename)
	defer C.free(unsafe.Pointer(cname))

	entry := C.git_tree_entry_byname(t.ptr, cname)
	if entry == nil {
		return nil
	}

	return newTreeEntry(entry)
}

func (t Tree) EntryByIndex(index uint64) *TreeEntry {
	entry := C.git_tree_entry_byindex(t.ptr, C.size_t(index))
	if entry == nil {
		return nil
	}

	return newTreeEntry(entry)
}

func (t Tree) EntryCount() uint64 {
	num := C.git_tree_entrycount(t.ptr)
	return uint64(num)
}

type TreeWalkCallback func(string, *TreeEntry) int

//export CallbackGitTreeWalk
func CallbackGitTreeWalk(_root unsafe.Pointer, _entry unsafe.Pointer, ptr unsafe.Pointer) C.int {
	root := C.GoString((*C.char)(_root))
	entry := (*C.git_tree_entry)(_entry)
	callback := *(*TreeWalkCallback)(ptr)

	return C.int(callback(root, newTreeEntry(entry)))
}

func (t Tree) Walk(callback TreeWalkCallback) error {
	err := C._go_git_treewalk(
		t.ptr,
		C.GIT_TREEWALK_PRE,
		unsafe.Pointer(&callback),
	)

	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

type TreeBuilder struct {
	ptr *C.git_treebuilder
	repo *Repository
}

func (v *TreeBuilder) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_treebuilder_free(v.ptr)
}

func (v *TreeBuilder) Insert(filename string, id *Oid, filemode int) (error) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	err := C.git_treebuilder_insert(nil, v.ptr, cfilename, id.toC(), C.git_filemode_t(filemode))
	if err < 0 {
		return MakeGitError(err)
	}

	return nil
}

func (v *TreeBuilder) Write() (*Oid, error) {
	oid := new(Oid)
	err := C.git_treebuilder_write(oid.toC(), v.repo.ptr, v.ptr)

	if err < 0 {
		return nil, MakeGitError(err)
	}

	return oid, nil
}
