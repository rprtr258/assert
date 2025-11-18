package src

// Creates a diff between old and new with the given algorithm capturing the ops.
//
// This is like [`diff`](crate::algorithms::diff) but instead of using an
// arbitrary hook this will always use [`Compact`] + [`Replace`] + [`Capture`]
// and return the captured [`DiffOp`]s.
func CaptureDiff[O comparable](
	alg Algorithm,
	old []O, old_range Range,
	new []O, new_range Range,
) Vec[DiffOp] {
	return CaptureDiffDeadline(alg, old, old_range, new, new_range, None[Instant]())
}

// Creates a diff between old and new with the given algorithm capturing the ops.
//
// Works like [`capture_diff`] but with an optional deadline.
func CaptureDiffDeadline[O comparable](
	alg Algorithm,
	old []O, old_range Range,
	new []O, new_range Range,
	deadline Option[Instant],
) Vec[DiffOp] {
	d := NewCompact(NewReplace(NewCapture()), old, new)
	_ = DiffDeadline(alg, &d, old, old_range, new, new_range, deadline).unwrap()
	return Vec[DiffOp](*d.d.d)
}
