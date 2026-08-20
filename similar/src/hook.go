package src

type usize = Usize

// A trait for reacting to an edit script from the "old" version to the "new" version.
type DiffHook interface {
	// Called when lines with indices `old_index` (in the old version) and
	// `new_index` (in the new version) start an section equal in both
	// versions, of length `len`.
	equal(old_index, new_index, len usize) Result[Unit]

	// Called when a section of length `old_len`, starting at `old_index`,
	// needs to be deleted from the old version.
	delete(old_index, old_len, new_index usize) Result[Unit]

	// Called when a section of the new version, of length `new_len`
	// and starting at `new_index`, needs to be inserted at position `old_index'.
	insert(old_index, new_index, new_len usize) Result[Unit]

	// Called when a section of the old version, starting at index
	// `old_index` and of length `old_len`, needs to be replaced with a
	// section of length `new_len`, starting at `new_index`, of the new
	// version.
	//
	// The default implementations invokes `delete` and `insert`.
	//
	// You can use the [`Replace`](crate::algorithms::Replace) hook to
	// automatically generate these.
	replace(old_index, old_len, new_index, new_len usize) Result[Unit]

	// Always called at the end of the algorithm.
	finish() Result[Unit]
}

type diffHookDefaultImpl struct{}

// Called when lines with indices `old_index` (in the old version) and
// `new_index` (in the new version) start an section equal in both
// versions, of length `len`.
func (diffHookDefaultImpl) equal(old_index, new_index, len usize) Result[Unit] {
	return Ok(Unit{})
}

// Called when a section of length `old_len`, starting at `old_index`,
// needs to be deleted from the old version.
func (diffHookDefaultImpl) delete(old_index, old_len, new_index usize) Result[Unit] {
	return Ok(Unit{})
}

// Called when a section of the new version, of length `new_len`
// and starting at `new_index`, needs to be inserted at position `old_index'.
func (diffHookDefaultImpl) insert(old_index, new_index, new_len usize) Result[Unit] {
	return Ok(Unit{})
}

// Called when a section of the old version, starting at index
// `old_index` and of length `old_len`, needs to be replaced with a
// section of length `new_len`, starting at `new_index`, of the new
// version.
//
// The default implementations invokes `delete` and `insert`.
//
// You can use the [`Replace`](crate::algorithms::Replace) hook to
// automatically generate these.
func (self diffHookDefaultImpl) replace(old_index, old_len, new_index, new_len usize) Result[Unit] {
	tmp := self.delete(old_index, old_len, new_index)
	if !tmp.Ok {
		return tmp
	}
	return self.insert(old_index, new_index, new_len)
}

// Always called at the end of the algorithm.
func (diffHookDefaultImpl) finish() Result[Unit] {
	return Ok(Unit{})
}

// Wrapper [`DiffHook`] that prevents calls to [`DiffHook::finish`].
//
// This hook is useful in situations where diff hooks are composed but you
// want to prevent that the finish hook method is called.
type NoFinishHook[D DiffHook] struct{ D D }

func newNoFinishHook[D DiffHook](d D) NoFinishHook[D] {
	return NoFinishHook[D]{d}
}

// Extracts the inner hook.
func (self NoFinishHook[D]) into_inner() D {
	return self.D
}

func (self NoFinishHook[D]) equal(old_index, new_index, len usize) Result[Unit] {
	return self.D.equal(old_index, new_index, len)
}

func (self NoFinishHook[D]) delete(old_index, old_len, new_index usize) Result[Unit] {
	return self.D.delete(old_index, old_len, new_index)
}

func (self NoFinishHook[D]) insert(old_index, new_index, new_len usize) Result[Unit] {
	return self.D.insert(old_index, new_index, new_len)
}

func (self NoFinishHook[D]) replace(old_index, old_len, new_index, new_len usize) Result[Unit] {
	return self.D.replace(old_index, old_len, new_index, new_len)
}

func (NoFinishHook[D]) finish() Result[Unit] {
	return Ok(Unit{})
}
