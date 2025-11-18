package src

import "slices"

// use super::utils::{common_prefix_len, common_suffix_len};
// use super::DiffHook;

// Performs semantic cleanup operations on a diff.
//
// This merges similar ops together but also tries to move hunks up and
// down the diff with the desire to connect as many hunks as possible.
// It still needs to be combined with [`Replace`](crate::algorithms::Replace)
// to get actual replace diff ops out.
type Compact[O comparable, D DiffHook] struct {
	d        D
	ops      Vec[DiffOp]
	old, new []O
}

// Creates a new compact hook wrapping another hook.
func NewCompact[O comparable, D DiffHook](d D, old, new []O) Compact[O, D] {
	return Compact[O, D]{d, nil, old, new}
}

func (self *Compact[O, D]) equal(old_index, new_index, len usize) Result[Unit] {
	self.ops = append(self.ops, DiffOp{
		DiffTagEqual,
		old_index,
		new_index,
		len,
		len,
	})
	return Ok(Unit{})
}

func (self *Compact[O, D]) delete(old_index, old_len, new_index usize) Result[Unit] {
	self.ops = append(self.ops, DiffOp{
		DiffTagDelete,
		old_index,
		new_index,
		old_len,
		0,
	})
	return Ok(Unit{})
}

func (self *Compact[O, D]) replace(old_index, old_len, new_index, new_len usize) Result[Unit] {
	tmp := self.delete(old_index, old_len, new_index)
	if !tmp.Ok {
		return tmp
	}
	return self.insert(old_index, new_index, new_len)
}

func (self *Compact[O, D]) insert(old_index, new_index, new_len usize) Result[Unit] {
	self.ops = append(self.ops, DiffOp{
		DiffTagInsert,
		old_index,
		new_index,
		0,
		new_len,
	})
	return Ok(Unit{})
}

func (self *Compact[O, D]) finish() Result[Unit] {
	cleanup_diff_ops(self.old, self.new, &self.ops)
	for _, op := range self.ops {
		if r := op.apply_to_hook(self.d); !r.Ok {
			return r
		}
	}
	return self.d.finish()
}

// Walks through all edits and shifts them up and then down, trying to see if
// they run into similar edits which can be merged.
func cleanup_diff_ops[O comparable](old, new []O, ops *Vec[DiffOp]) {
	// First attempt to compact all Deletions
	pointer := Usize(0)
	for pointer < ops.len() {
		op := (*ops)[pointer]
		if op.DiffTag == DiffTagDelete {
			pointer = shift_diff_ops_up(ops, old, new, &pointer)
			pointer = shift_diff_ops_down(ops, old, new, &pointer)
		}
		pointer++
	}

	// Then attempt to compact all Insertions
	pointer = 0
	for pointer < ops.len() {
		op := (*ops)[pointer]
		if op.DiffTag == DiffTagInsert {
			pointer = shift_diff_ops_up(ops, old, new, &pointer)
			pointer = shift_diff_ops_down(ops, old, new, &pointer)
		}
		pointer++
	}
}

func shift_diff_ops_up[O comparable](
	ops *Vec[DiffOp],
	old, new []O,
	pointer *usize,
) usize {
	for {
		prev_op, ok := AndThen(pointer.CheckedSub(1), ops.Get).Unpack()
		if !ok {
			break
		}
		this_op := (*ops)[*pointer]
		switch [2]DiffTag{this_op.DiffTag, prev_op.DiffTag} {
		// Shift Inserts Upwards
		case [2]DiffTag{DiffTagInsert, DiffTagEqual}:
			suffix_len := common_suffix_len(old, prev_op.old_range(), new, this_op.new_range())
			if suffix_len > 0 {
				if tag, ok := Map(ops.Get(*pointer+1), func(x DiffOp) DiffTag { return x.DiffTag }).Unpack(); ok && tag == DiffTagEqual {
					(*ops)[*pointer+1].grow_left(suffix_len)
				} else {
					ops.insert(
						*pointer+1,
						DiffOp{
							DiffTagEqual,
							prev_op.old_range().end - suffix_len,
							this_op.new_range().end - suffix_len,
							suffix_len,
							suffix_len,
						},
					)
				}
				(*ops)[*pointer].shift_left(suffix_len)
				(*ops)[*pointer-1].shrink_left(suffix_len)

				if (*ops)[*pointer-1].is_empty() {
					ops.remove(*pointer - 1)
					*pointer -= 1
				}
			} else if (*ops)[*pointer-1].is_empty() {
				ops.remove(*pointer - 1)
				*pointer -= 1
			} else {
				// We can't shift upwards anymore
				break
			}
		// Shift Deletions Upwards
		case [2]DiffTag{DiffTagDelete, DiffTagEqual}:
			// check common suffix for the amount we can shift
			suffix_len := common_suffix_len(old, prev_op.old_range(), new, this_op.new_range())
			if suffix_len != 0 {
				if tag, ok := Map(ops.Get(*pointer+1), func(x DiffOp) DiffTag { return x.DiffTag }).Unpack(); ok && tag == DiffTagEqual {
					(*ops)[*pointer+1].grow_left(suffix_len)
				} else {
					old_range := prev_op.old_range()
					*ops = slices.Insert(*ops,
						int(*pointer+1),
						DiffOp{
							DiffTagEqual,
							old_range.end - suffix_len,
							this_op.new_range().end - suffix_len,
							old_range.len() - suffix_len,
							old_range.len() - suffix_len,
						},
					)
				}
				(*ops)[*pointer].shift_left(suffix_len)
				(*ops)[*pointer-1].shrink_left(suffix_len)

				if (*ops)[*pointer-1].is_empty() {
					ops.remove(*pointer - 1)
					*pointer -= 1
				}
			} else if (*ops)[*pointer-1].is_empty() {
				ops.remove(*pointer - 1)
				*pointer -= 1
			} else {
				// We can't shift upwards anymore
				break
			}
		// Swap the Delete and Insert
		case [2]DiffTag{DiffTagInsert, DiffTagDelete}, [2]DiffTag{DiffTagDelete, DiffTagInsert}:
			ops.swap(*pointer-1, *pointer)
			*pointer -= 1
			// Merge the two ranges
		case [2]DiffTag{DiffTagInsert, DiffTagInsert}:
			(*ops)[*pointer-1].grow_right(this_op.new_range().len())
			ops.remove(*pointer)
			*pointer -= 1
		case [2]DiffTag{DiffTagDelete, DiffTagDelete}:
			(*ops)[*pointer-1].grow_right(this_op.old_range().len())
			ops.remove(*pointer)
			*pointer -= 1
		default:
			panic("unexpected tag")
		}
	}
	return *pointer
}

func shift_diff_ops_down[O comparable](
	ops *Vec[DiffOp],
	old, new []O,
	pointer *usize,
) usize {
	for {
		next_op, ok := AndThen(pointer.CheckedSub(1), ops.Get).Unpack()
		if !ok {
			break
		}
		this_op := (*ops)[*pointer]
		switch [2]DiffTag{this_op.DiffTag, next_op.DiffTag} {
		// Shift Inserts Downwards
		case [2]DiffTag{DiffTagInsert, DiffTagEqual}:
			prefix_len := common_prefix_len(old, next_op.old_range(), new, this_op.new_range())
			if prefix_len > 0 {
				if tag, ok := Map(AndThen(pointer.CheckedSub(1), ops.Get), func(x DiffOp) DiffTag { return x.DiffTag }).Unpack(); ok && tag == DiffTagEqual {
					(*ops)[*pointer-1].grow_right(prefix_len)
				} else {
					ops.insert(
						*pointer,
						DiffOp{
							DiffTagEqual,
							next_op.old_range().start,
							this_op.new_range().start,
							prefix_len,
							prefix_len,
						},
					)
					*pointer += 1
				}
				(*ops)[*pointer].shift_right(prefix_len)
				(*ops)[*pointer+1].shrink_right(prefix_len)

				if (*ops)[*pointer+1].is_empty() {
					ops.remove(*pointer + 1)
				}
			} else if (*ops)[*pointer+1].is_empty() {
				ops.remove(*pointer + 1)
			} else {
				// We can't shift upwards anymore
				break
			}
			// Shift Deletions Downwards
		case [2]DiffTag{DiffTagDelete, DiffTagEqual}:
			// check common suffix for the amount we can shift
			prefix_len := common_prefix_len(old, next_op.old_range(), new, this_op.new_range())
			if prefix_len > 0 {
				if tag, ok := Map(AndThen(pointer.CheckedSub(1), ops.Get), func(x DiffOp) DiffTag { return x.DiffTag }).Unpack(); ok && tag == DiffTagEqual {
					(*ops)[*pointer-1].grow_right(prefix_len)
				} else {
					ops.insert(
						*pointer,
						DiffOp{
							DiffTagEqual,
							next_op.old_range().start,
							this_op.new_range().start,
							prefix_len,
							prefix_len,
						},
					)
					*pointer += 1
				}
				(*ops)[*pointer].shift_right(prefix_len)
				(*ops)[*pointer+1].shrink_right(prefix_len)

				if (*ops)[*pointer+1].is_empty() {
					ops.remove(*pointer + 1)
				}
			} else if (*ops)[*pointer+1].is_empty() {
				ops.remove(*pointer + 1)
			} else {
				// We can't shift downwards anymore
				break
			}
			// Swap the Delete and Insert
		case [2]DiffTag{DiffTagInsert, DiffTagDelete}, [2]DiffTag{DiffTagDelete, DiffTagInsert}:
			ops.swap(*pointer, *pointer+1)
			*pointer += 1
			// Merge the two ranges
		case [2]DiffTag{DiffTagInsert, DiffTagInsert}:
			(*ops)[*pointer].grow_right(next_op.new_range().len())
			ops.remove(*pointer + 1)
		case [2]DiffTag{DiffTagDelete, DiffTagDelete}:
			(*ops)[*pointer].grow_right(next_op.old_range().len())
			ops.remove(*pointer + 1)
		default:
			panic("unexpected tag")
		}
	}

	return *pointer
}
