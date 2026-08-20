package src

type Algorithm uint8

const (
	// Picks the myers algorithm from [`crate::algorithms::myers`]
	AlgorithmMyers Algorithm = iota
	// Picks the patience algorithm from [`crate::algorithms::patience`]
	AlgorithmPatience
	// Picks the LCS algorithm from [`crate::algorithms::lcs`]
	AlgorithmLcs
)

type DiffTag uint8

const (
	// A segment is equal (see [`DiffHook::equal`])
	DiffTagEqual DiffTag = iota
	// A segment was deleted (see [`DiffHook::delete`])
	DiffTagDelete
	// A segment was inserted (see [`DiffHook::insert`])
	DiffTagInsert
	// A segment was replaced (see [`DiffHook::replace`])
	DiffTagReplace
)

// Utility enum to capture a diff operation.
// This is used by [`Capture`](crate::algorithms::Capture).
type DiffOp struct {
	DiffTag
	// The starting index in the old sequence.
	old_index Usize
	// The starting index in the new sequence.
	new_index Usize
	// TODO: same, if Equal
	// The length of the old segment.
	old_len Usize
	// The length of the new segment.
	new_len Usize
}

// Transform the op into a tuple of diff tag and ranges.
//
// This is useful when operating on slices.  The returned format is
// `(tag, i1..i2, j1..j2)`:
//
// * `Replace`: `a[i1..i2]` should be replaced by `b[j1..j2]`
// * `Delete`: `a[i1..i2]` should be deleted (`j1 == j2` in this case).
// * `Insert`: `b[j1..j2]` should be inserted at `a[i1..i2]` (`i1 == i2` in this case).
// * `Equal`: `a[i1..i2]` is equal to `b[j1..j2]`.
func (self DiffOp) as_tag_tuple() (Range, Range) {
	old_index := self.old_index
	new_index := self.new_index
	old_len := self.old_len
	new_len := self.new_len
	switch self.DiffTag {
	case DiffTagEqual:
		len := old_len // == self.new_len
		return Range{old_index, old_index + len},
			Range{new_index, new_index + len}
	case DiffTagDelete:
		return Range{old_index, old_index + old_len},
			Range{new_index, new_index}
	case DiffTagInsert:
		return Range{old_index, old_index},
			Range{new_index, new_index + new_len}
	case DiffTagReplace:
		return Range{old_index, old_index + old_len},
			Range{new_index, new_index + new_len}
	default:
		panic("unreachable")
	}
}

// Returns the old range.
func (self DiffOp) old_range() Range {
	res, _ := self.as_tag_tuple()
	return res
}

// Returns the new range.
func (self DiffOp) new_range() Range {
	_, res := self.as_tag_tuple()
	return res
}

// Apply this operation to a diff hook.
func (self DiffOp) apply_to_hook(d DiffHook) Result[Unit] {
	switch self.DiffTag {
	case DiffTagEqual:
		len := self.old_len // TODO: same as self.new_len
		return d.equal(self.old_index, self.new_index, len)
	case DiffTagDelete:
		return d.delete(self.old_index, self.old_len, self.new_index)
	case DiffTagInsert:
		return d.insert(self.old_index, self.new_index, self.new_len)
	case DiffTagReplace:
		return d.replace(self.old_index, self.old_len, self.new_index, self.new_len)
	default:
		panic("unreachable")
	}
}

func (self DiffOp) is_empty() bool {
	old, new := self.as_tag_tuple()
	return is_empty_range(old) && is_empty_range(new)
}

func (self *DiffOp) shift_left(adjust usize) {
	self.adjust(adjust, true, 0, false)
}

func (self *DiffOp) shift_right(adjust usize) {
	self.adjust(adjust, false, 0, false)
}

func (self *DiffOp) grow_left(adjust usize) {
	self.adjust(adjust, true, adjust, false)
}

func (self *DiffOp) grow_right(adjust usize) {
	self.adjust(0, false, adjust, false)
}

func (self *DiffOp) shrink_left(adjust usize) {
	self.adjust(0, false, adjust, true)
}

func (self *DiffOp) shrink_right(adjust usize) {
	self.adjust(adjust, false, adjust, true)
}

func (self *DiffOp) adjust(
	adjust_offseta usize, adjust_offsetb bool,
	adjust_lena usize, adjust_lenb bool,
) {
	modify := func(val *usize, adja usize, adjb bool) {
		if adjb {
			*val -= adja
		} else {
			*val += adja
		}
	}

	old_index := &self.old_index
	new_index := &self.new_index
	old_len := &self.old_len
	new_len := &self.new_len
	switch self.DiffTag {
	case DiffTagEqual:
		modify(old_index, adjust_offseta, adjust_offsetb)
		modify(new_index, adjust_offseta, adjust_offsetb)
		len := old_len // == self.new_len
		modify(len, adjust_lena, adjust_lenb)
		*new_len = *len
	case DiffTagDelete:
		modify(old_index, adjust_offseta, adjust_offsetb)
		modify(old_len, adjust_lena, adjust_lenb)
		modify(new_index, adjust_offseta, adjust_offsetb)
	case DiffTagInsert:
		modify(old_index, adjust_offseta, adjust_offsetb)
		modify(new_index, adjust_offseta, adjust_offsetb)
		modify(new_len, adjust_lena, adjust_lenb)
	case DiffTagReplace:
		modify(old_index, adjust_offseta, adjust_offsetb)
		modify(old_len, adjust_lena, adjust_lenb)
		modify(new_index, adjust_offseta, adjust_offsetb)
		modify(new_len, adjust_lena, adjust_lenb)
	default:
		panic("unreachable")
	}
}
