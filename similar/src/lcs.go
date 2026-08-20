package src

// //! LCS diff algorithm.
// //!
// //! * time: `O((NM)D log (M)D)`
// //! * space `O(MN)`
// use std::collections::BTreeMap;
// use std::ops::{Index, Range};

// use crate::algorithms::utils::{common_prefix_len, common_suffix_len, is_empty_range};
// use crate::algorithms::DiffHook;
// use crate::deadline_support::{deadline_exceeded, Instant};

// /// LCS diff algorithm.
// ///
// /// Diff `old`, between indices `old_range` and `new` between indices `new_range`.
// ///
// /// This diff is done with an optional deadline that defines the maximal
// /// execution time permitted before it bails and falls back to an very bad
// /// approximation.  Deadlines with LCS do not make a lot of sense and should
// /// not be used.
// pub fn diff<Old, New, D>(
//     d: &mut D,
//     old: &Old,
//     old_range: Range<usize>,
//     new: &New,
//     new_range: Range<usize>,
// ) -> Result<(), D::Error>
// where
//     Old: Index<usize> + ?Sized,
//     New: Index<usize> + ?Sized,
//     D: DiffHook,
//     New::Output: PartialEq<Old::Output>,
// {
//     diff_deadline(d, old, old_range, new, new_range, None)
// }

// LCS diff algorithm.
//
// Diff `old`, between indices `old_range` and `new` between indices `new_range`.
//
// This diff is done with an optional deadline that defines the maximal
// execution time permitted before it bails and falls back to an approximation.
func lcs_diff_deadline[O comparable, D DiffHook](
	d D,
	old []O, old_range Range,
	new []O, new_range Range,
	deadline Option[Instant],
) Result[Unit] {
	if is_empty_range(new_range) {
		if tmp := d.delete(old_range.start, old_range.len(), new_range.start); !tmp.Ok {
			return tmp
		}
		return d.finish()
	} else if is_empty_range(old_range) {
		if tmp := d.insert(old_range.start, new_range.start, new_range.len()); !tmp.Ok {
			return tmp
		}
		return d.finish()
	}

	common_prefix_len := common_prefix_len(old, old_range, new, new_range)
	common_suffix_len := common_suffix_len(
		old,
		Range{old_range.start + common_prefix_len, old_range.end},
		new,
		Range{new_range.start + common_prefix_len, new_range.end},
	)

	// If the sequences are not different then we're done
	if common_prefix_len == old_range.len() && (old_range.len() == new_range.len()) {
		if tmp := d.equal(0, 0, old_range.len()); !tmp.Ok {
			return tmp
		}
		return d.finish()
	}

	maybe_table := make_table(
		old,
		Range{common_prefix_len, old_range.len() - common_suffix_len},
		new,
		Range{common_prefix_len, new_range.len() - common_suffix_len},
		deadline,
	)
	old_idx := usize(0)
	new_idx := usize(0)
	new_len := new_range.len() - common_prefix_len - common_suffix_len
	old_len := old_range.len() - common_prefix_len - common_suffix_len

	if common_prefix_len > 0 {
		if tmp := d.equal(old_range.start, new_range.start, common_prefix_len); !tmp.Ok {
			return tmp
		}
	}

	if table, ok := maybe_table.Unpack(); ok {
		for new_idx < new_len && old_idx < old_len {
			old_orig_idx := old_range.start + common_prefix_len + old_idx
			new_orig_idx := new_range.start + common_prefix_len + new_idx

			if new[new_orig_idx] == old[old_orig_idx] {
				if tmp := d.equal(old_orig_idx, new_orig_idx, 1); !tmp.Ok {
					return tmp
				}
				old_idx += 1
				new_idx += 1
			} else if table[[2]usize{new_idx, old_idx + 1}] >= table[[2]usize{new_idx + 1, old_idx}] {
				if tmp := d.delete(old_orig_idx, 1, new_orig_idx); !tmp.Ok {
					return tmp
				}
				old_idx += 1
			} else {
				if tmp := d.insert(old_orig_idx, new_orig_idx, 1); !tmp.Ok {
					return tmp
				}
				new_idx += 1
			}
		}
	} else {
		old_orig_idx := old_range.start + common_prefix_len + old_idx
		new_orig_idx := new_range.start + common_prefix_len + new_idx
		if tmp := d.delete(old_orig_idx, old_len, new_orig_idx); !tmp.Ok {
			return tmp
		}
		if tmp := d.insert(old_orig_idx, new_orig_idx, new_len); !tmp.Ok {
			return tmp
		}
	}

	if old_idx < old_len {
		if tmp := d.delete(
			old_range.start+common_prefix_len+old_idx,
			old_len-old_idx,
			new_range.start+common_prefix_len+new_idx,
		); !tmp.Ok {
			return tmp
		}
		old_idx += old_len - old_idx
	}

	if new_idx < new_len {
		if tmp := d.insert(
			old_range.start+common_prefix_len+old_idx,
			new_range.start+common_prefix_len+new_idx,
			new_len-new_idx,
		); !tmp.Ok {
			return tmp
		}
	}

	if common_suffix_len > 0 {
		if tmp := d.equal(
			old_range.start+old_len+common_prefix_len,
			new_range.start+new_len+common_prefix_len,
			common_suffix_len,
		); !tmp.Ok {
			return tmp
		}
	}

	return d.finish()
}

func make_table[O comparable](
	old []O, old_range Range,
	new []O, new_range Range,
	deadline Option[Instant],
) Option[map[[2]usize]uint32] {
	old_len := old_range.len()
	new_len := new_range.len()
	table := map[[2]usize]uint32{}

	for i0 := range new_len {
		i := new_len - i0 - 1
		// are we running for too long?  give up on the table
		if deadline_exceeded(deadline) {
			return None[map[[2]usize]uint32]()
		}

		for j0 := range new_len {
			j := old_len - j0 - 1
			var val uint32
			if new[i] == old[j] {
				val = table[[2]usize{i + 1, j + 1}] + 1
			} else {
				val = max(table[[2]usize{i + 1, j}], table[[2]usize{i, j + 1}])
			}
			if val > 0 {
				table[[2]usize{i, j}] = val
			}
		}
	}

	return Some(table)
}

// #[test]
// fn test_table() {
//     let table = make_table(&vec![2, 3], 0..2, &vec![0, 1, 2], 0..3, None).unwrap();
//     let expected = {
//         let mut m = BTreeMap::new();
//         m.insert((1, 0), 1);
//         m.insert((0, 0), 1);
//         m.insert((2, 0), 1);
//         m
//     };
//     assert_eq!(table, expected);
// }

// #[test]
// fn test_diff() {
//     let a: &[usize] = &[0, 1, 2, 3, 4];
//     let b: &[usize] = &[0, 1, 2, 9, 4];

//     let mut d = crate::algorithms::Replace::new(crate::algorithms::Capture::new());
//     diff(&mut d, a, 0..a.len(), b, 0..b.len()).unwrap();
//     insta::assert_debug_snapshot!(d.into_inner().ops());
// }

// #[test]
// fn test_contiguous() {
//     let a: &[usize] = &[0, 1, 2, 3, 4, 4, 4, 5];
//     let b: &[usize] = &[0, 1, 2, 8, 9, 4, 4, 7];

//     let mut d = crate::algorithms::Replace::new(crate::algorithms::Capture::new());
//     diff(&mut d, a, 0..a.len(), b, 0..b.len()).unwrap();
//     insta::assert_debug_snapshot!(d.into_inner().ops());
// }

// #[test]
// fn test_pat() {
//     let a: &[usize] = &[0, 1, 3, 4, 5];
//     let b: &[usize] = &[0, 1, 4, 5, 8, 9];

//     let mut d = crate::algorithms::Capture::new();
//     diff(&mut d, a, 0..a.len(), b, 0..b.len()).unwrap();
//     insta::assert_debug_snapshot!(d.ops());
// }

// #[test]
// fn test_same() {
//     let a: &[usize] = &[0, 1, 2, 3, 4, 4, 4, 5];
//     let b: &[usize] = &[0, 1, 2, 3, 4, 4, 4, 5];

//     let mut d = crate::algorithms::Capture::new();
//     diff(&mut d, a, 0..a.len(), b, 0..b.len()).unwrap();
//     insta::assert_debug_snapshot!(d.ops());
// }

// #[test]
// fn test_finish_called() {
//     struct HasRunFinish(bool);

//     impl DiffHook for HasRunFinish {
//         type Error = ();
//         fn finish(&mut self) -> Result<(), Self::Error> {
//             self.0 = true;
//             Ok(())
//         }
//     }

//     let mut d = HasRunFinish(false);
//     let slice = &[1, 2];
//     let slice2 = &[1, 2, 3];
//     diff(&mut d, slice, 0..slice.len(), slice2, 0..slice2.len()).unwrap();
//     assert!(d.0);

//     let mut d = HasRunFinish(false);
//     let slice = &[1, 2];
//     diff(&mut d, slice, 0..slice.len(), slice, 0..slice.len()).unwrap();
//     assert!(d.0);

//     let mut d = HasRunFinish(false);
//     let slice: &[u8] = &[];
//     diff(&mut d, slice, 0..slice.len(), slice, 0..slice.len()).unwrap();
//     assert!(d.0);
// }

// #[test]
// fn test_bad_range_regression() {
//     use crate::algorithms::Capture;
//     use crate::DiffOp;
//     let mut d = Capture::new();
//     diff(&mut d, &[0], 0..1, &[0, 0], 0..2).unwrap();
//     assert_eq!(
//         d.into_ops(),
//         vec![
//             DiffOp::Equal {
//                 old_index: 0,
//                 new_index: 0,
//                 len: 1
//             },
//             DiffOp::Insert {
//                 old_index: 1,
//                 new_index: 1,
//                 new_len: 1
//             }
//         ]
//     );
// }
