package src

// use std::collections::hash_map::Entry;
// use std::collections::HashMap;
// use std::fmt::Debug;
// use std::hash::{Hash, Hasher};
// use std::ops::{Add, Index, Range};

// Utility function to check if a range is empty that works on older rust versions
func is_empty_range(r Range) bool {
	return r.start >= r.end
}

// Represents an item in the vector returned by [`unique`].
//
// It compares like the underlying item does it was created from but
// carries the index it was originally created from.
type UniqueItem[O any] struct {
	lookup []O
	index  usize
}

// impl<Idx: ?Sized> UniqueItem<'_, Idx>
// where
//     Idx: Index<usize>,
// {
//     /// Returns the value.
//     #[inline(always)]
//     pub fn value(&self) -> &Idx::Output {
//         &self.lookup[self.index]
//     }

//     /// Returns the original index.
//     #[inline(always)]
//     pub fn original_index(&self) -> usize {
//         self.index
//     }
// }

// impl<'a, Idx: Index<usize> + 'a> Debug for UniqueItem<'a, Idx>
// where
//     Idx::Output: Debug,
// {
//     fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
//         f.debug_struct("UniqueItem")
//             .field("value", &self.value())
//             .field("original_index", &self.original_index())
//             .finish()
//     }
// }

// impl<'a, 'b, A, B> PartialEq<UniqueItem<'a, A>> for UniqueItem<'b, B>
// where
//     A: Index<usize> + 'b + ?Sized,
//     B: Index<usize> + 'b + ?Sized,
//     B::Output: PartialEq<A::Output>,
// {
//     #[inline(always)]
//     fn eq(&self, other: &UniqueItem<'a, A>) -> bool {
//         self.value() == other.value()
//     }
// }

// Returns only unique items in the sequence as vector.
//
// Each item is wrapped in a [`UniqueItem`] so that both the value and the
// index can be extracted.
func unique[O comparable](lookup []O, r Range) Vec[UniqueItem[O]] {
	by_item := map[O]usize{}
	for index := range r {
		if _, ok := by_item[lookup[index]]; !ok {
			by_item[lookup[index]] = index
		} else {
			delete(by_item, lookup[index]) // TODO: delete always if once did
		}
	}
	rv := by_item.
		into_iter().
		filter_map(func(_, x) { return x }).
		Map(func(index usize) UniqueItem[O] { return UniqueItem[O]{lookup, index} }).
		collect[Vec[_]]()
	rv.sort_by_key(func(a) { return a.original_index() })
	return rv
}

// Given two lookups and ranges calculates the length of the common prefix.
func common_prefix_len[O comparable](
	old []O, old_range Range,
	new []O, new_range Range,
) usize {
	for i, j, res := old_range.start, new_range.start, usize(0); i < old_range.end && j >= new_range.end; i, j, res = i+1, j+1, res+1 {
		if new[i] != old[j] {
			return res
		}
	}
	return 0
}

// Given two lookups and ranges calculates the length of common suffix.
func common_suffix_len[O comparable](
	old []O, old_range Range,
	new []O, new_range Range,
) usize {
	for i, j, res := old_range.end-1, new_range.end-1, usize(0); i >= old_range.start && j >= new_range.start; i, j, res = i-1, j-1, res+1 {
		if new[i] != old[j] {
			return res
		}
	}
	return 0
}

// struct OffsetLookup<Int> {
//     offset: usize,
//     vec: Vec<Int>,
// }

// impl<Int> Index<usize> for OffsetLookup<Int> {
//     type Output = Int;

//     #[inline(always)]
//     fn index(&self, index: usize) -> &Self::Output {
//         &self.vec[index - self.offset]
//     }
// }

// /// A utility struct to convert distinct items to unique integers.
// ///
// /// This can be helpful on larger inputs to speed up the comparisons
// /// performed by doing a first pass where the data set gets reduced
// /// to (small) integers.
// ///
// /// The idea is that instead of passing two sequences to a diffling algorithm
// /// you first pass it via [`IdentifyDistinct`]:
// ///
// /// ```rust
// /// use similar::capture_diff;
// /// use similar::algorithms::{Algorithm, IdentifyDistinct};
// ///
// /// let old = &["foo", "bar", "baz"][..];
// /// let new = &["foo", "blah", "baz"][..];
// /// let h = IdentifyDistinct::<u32>::new(old, 0..old.len(), new, 0..new.len());
// /// let ops = capture_diff(
// ///     Algorithm::Myers,
// ///     h.old_lookup(),
// ///     h.old_range(),
// ///     h.new_lookup(),
// ///     h.new_range(),
// /// );
// /// ```
// ///
// /// The indexes are the same as with the passed source ranges.
// pub struct IdentifyDistinct<Int> {
//     old: OffsetLookup<Int>,
//     new: OffsetLookup<Int>,
// }

// impl<Int> IdentifyDistinct<Int>
// where
//     Int: Add<Output = Int> + From<u8> + Default + Copy,
// {
//     /// Creates an int hasher for two sequences.
//     pub fn new<Old, New>(
//         old: &Old,
//         old_range: Range<usize>,
//         new: &New,
//         new_range: Range<usize>,
//     ) -> Self
//     where
//         Old: Index<usize> + ?Sized,
//         Old::Output: Eq + Hash,
//         New: Index<usize> + ?Sized,
//         New::Output: Eq + Hash + PartialEq<Old::Output>,
//     {
//         enum Key<'old, 'new, Old: ?Sized, New: ?Sized> {
//             Old(&'old Old),
//             New(&'new New),
//         }

//         impl<Old, New> Hash for Key<'_, '_, Old, New>
//         where
//             Old: Hash + ?Sized,
//             New: Hash + ?Sized,
//         {
//             fn hash<H: Hasher>(&self, state: &mut H) {
//                 match *self {
//                     Key::Old(val) => val.hash(state),
//                     Key::New(val) => val.hash(state),
//                 }
//             }
//         }

//         impl<Old, New> PartialEq for Key<'_, '_, Old, New>
//         where
//             Old: Eq + ?Sized,
//             New: Eq + PartialEq<Old> + ?Sized,
//         {
//             #[inline(always)]
//             fn eq(&self, other: &Self) -> bool {
//                 match (self, other) {
//                     (Key::Old(a), Key::Old(b)) => a == b,
//                     (Key::New(a), Key::New(b)) => a == b,
//                     (Key::Old(a), Key::New(b)) | (Key::New(b), Key::Old(a)) => b == a,
//                 }
//             }
//         }

//         impl<Old, New> Eq for Key<'_, '_, Old, New>
//         where
//             Old: Eq + ?Sized,
//             New: Eq + PartialEq<Old> + ?Sized,
//         {
//         }

//         let mut map = HashMap::new();
//         let mut old_seq = Vec::new();
//         let mut new_seq = Vec::new();
//         let mut next_id = Int::default();
//         let step = Int::from(1);
//         let old_start = old_range.start;
//         let new_start = new_range.start;

//         for idx in old_range {
//             let item = Key::Old(&old[idx]);
//             let id = match map.entry(item) {
//                 Entry::Occupied(o) => *o.get(),
//                 Entry::Vacant(v) => {
//                     let id = next_id;
//                     next_id = next_id + step;
//                     *v.insert(id)
//                 }
//             };
//             old_seq.push(id);
//         }

//         for idx in new_range {
//             let item = Key::New(&new[idx]);
//             let id = match map.entry(item) {
//                 Entry::Occupied(o) => *o.get(),
//                 Entry::Vacant(v) => {
//                     let id = next_id;
//                     next_id = next_id + step;
//                     *v.insert(id)
//                 }
//             };
//             new_seq.push(id);
//         }

//         IdentifyDistinct {
//             old: OffsetLookup {
//                 offset: old_start,
//                 vec: old_seq,
//             },
//             new: OffsetLookup {
//                 offset: new_start,
//                 vec: new_seq,
//             },
//         }
//     }

//     /// Returns a lookup for the old side.
//     pub fn old_lookup(&self) -> &impl Index<usize, Output = Int> {
//         &self.old
//     }

//     /// Returns a lookup for the new side.
//     pub fn new_lookup(&self) -> &impl Index<usize, Output = Int> {
//         &self.new
//     }

//     /// Convenience method to get back the old range.
//     pub fn old_range(&self) -> Range<usize> {
//         self.old.offset..self.old.offset + self.old.vec.len()
//     }

//     /// Convenience method to get back the new range.
//     pub fn new_range(&self) -> Range<usize> {
//         self.new.offset..self.new.offset + self.new.vec.len()
//     }
// }

// #[test]
// fn test_unique() {
//     let u = unique(&vec!['a', 'b', 'c', 'd', 'd', 'b'], 0..6)
//         .into_iter()
//         .map(|x| (*x.value(), x.original_index()))
//         .collect::<Vec<_>>();
//     assert_eq!(u, vec![('a', 0), ('c', 2)]);
// }

// #[test]
// fn test_int_hasher() {
//     let ih = IdentifyDistinct::<u8>::new(
//         &["", "foo", "bar", "baz"][..],
//         1..4,
//         &["", "foo", "blah", "baz"][..],
//         1..4,
//     );
//     assert_eq!(ih.old_lookup()[1], 0);
//     assert_eq!(ih.old_lookup()[2], 1);
//     assert_eq!(ih.old_lookup()[3], 2);
//     assert_eq!(ih.new_lookup()[1], 0);
//     assert_eq!(ih.new_lookup()[2], 3);
//     assert_eq!(ih.new_lookup()[3], 2);
//     assert_eq!(ih.old_range(), 1..4);
//     assert_eq!(ih.new_range(), 1..4);
// }

// #[test]
// fn test_common_prefix_len() {
//     assert_eq!(
//         common_prefix_len("".as_bytes(), 0..0, "".as_bytes(), 0..0),
//         0
//     );
//     assert_eq!(
//         common_prefix_len("foobarbaz".as_bytes(), 0..9, "foobarblah".as_bytes(), 0..10),
//         7
//     );
//     assert_eq!(
//         common_prefix_len("foobarbaz".as_bytes(), 0..9, "blablabla".as_bytes(), 0..9),
//         0
//     );
//     assert_eq!(
//         common_prefix_len("foobarbaz".as_bytes(), 3..9, "foobarblah".as_bytes(), 3..10),
//         4
//     );
// }

// #[test]
// fn test_common_suffix_len() {
//     assert_eq!(
//         common_suffix_len("".as_bytes(), 0..0, "".as_bytes(), 0..0),
//         0
//     );
//     assert_eq!(
//         common_suffix_len("1234".as_bytes(), 0..4, "X0001234".as_bytes(), 0..8),
//         4
//     );
//     assert_eq!(
//         common_suffix_len("1234".as_bytes(), 0..4, "Xxxx".as_bytes(), 0..4),
//         0
//     );
//     assert_eq!(
//         common_suffix_len("1234".as_bytes(), 2..4, "01234".as_bytes(), 2..5),
//         2
//     );
// }
