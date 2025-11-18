package src

// //! Myers' diff algorithm.
// //!
// //! * time: `O((N+M)D)`
// //! * space `O(N+M)`
// //!
// //! See [the original article by Eugene W. Myers](http://www.xmailserver.org/diff2.pdf)
// //! describing it.
// //!
// //! The implementation of this algorithm is based on the implementation by
// //! Brandon Williams.
// //!
// //! # Heuristics
// //!
// //! At present this implementation of Myers' does not implement any more advanced
// //! heuristics that would solve some pathological cases.  For instance passing two
// //! large and completely distinct sequences to the algorithm will make it spin
// //! without making reasonable progress.  Currently the only protection in the
// //! library against this is to pass a deadline to the diffing algorithm.
// //!
// //! For potential improvements here see [similar#15](https://github.com/mitsuhiko/similar/issues/15).

// use std::ops::{Index, IndexMut, Range};

// use crate::algorithms::utils::{common_prefix_len, common_suffix_len, is_empty_range};
// use crate::algorithms::DiffHook;
// use crate::deadline_support::{deadline_exceeded, Instant};

// /// Myers' diff algorithm.
// ///
// /// Diff `old`, between indices `old_range` and `new` between indices `new_range`.
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

// Myers' diff algorithm with deadline.
//
// Diff `old`, between indices `old_range` and `new` between indices `new_range`.
//
// This diff is done with an optional deadline that defines the maximal
// execution time permitted before it bails and falls back to an approximation.
func myers_diff_deadline[O comparable, D DiffHook](
	d D,
	old []O, old_range Range,
	new []O, new_range Range,
	deadline Option[Instant],
) Result[Unit] {
	max_d := max_d(old_range.len(), new_range.len())
	vb := newV(max_d)
	vf := newV(max_d)
	if tmp := conquer(d, old, old_range, new, new_range, &vf, &vb, deadline); !tmp.Ok {
		return tmp
	}
	return d.finish()
}

// // A D-path is a path which starts at (0,0) that has exactly D non-diagonal
// // edges. All D-paths consist of a (D - 1)-path followed by a non-diagonal edge
// // and then a possibly empty sequence of diagonal edges called a snake.

// `V` contains the endpoints of the furthest reaching `D-paths`. For each
// recorded endpoint `(x,y)` in diagonal `k`, we only need to retain `x` because
// `y` can be computed from `x - k`. In other words, `V` is an array of integers
// where `V[k]` contains the row index of the endpoint of the furthest reaching
// path in diagonal `k`.
//
// We can't use a traditional Vec to represent `V` since we use `k` as an index
// and it can take on negative values. So instead `V` is represented as a
// light-weight wrapper around a Vec plus an `offset` which is the maximum value
// `k` can take on in order to map negative `k`'s back to a value >= 0.
type V struct {
	offset usize
	v      Vec[usize] // Look into initializing this to -1 and storing isize
}

func newV(max_d usize) V {
	return V{
		offset: max_d,
		v:      make(Vec[usize], 0, 2*max_d),
	}
}

func (self V) len() usize {
	return self.v.len()
}

func (self V) get(index usize) usize {
	return self.v[index+self.offset]
}

func (self V) set(index usize, value usize) {
	self.v[index+self.offset] = value
}

func max_d(len1, len2 usize) usize {
	// XXX look into reducing the need to have the additional '+ 1'
	return (len1+len2+1)/2 + 1
}

func split_at(r Range, at usize) (Range, Range) {
	return Range{r.start, at}, Range{at, r.end}
}

// / A `Snake` is a sequence of diagonal edges in the edit graph.  Normally
// / a snake has a start end end point (and it is possible for a snake to have
// / a length of zero, meaning the start and end points are the same) however
// / we do not need the end point which is why it's not implemented here.
// /
// / The divide part of a divide-and-conquer strategy. A D-path has D+1 snakes
// / some of which may be empty. The divide step requires finding the ceil(D/2) +
// / 1 or middle snake of an optimal D-path. The idea for doing so is to
// / simultaneously run the basic algorithm in both the forward and reverse
// / directions until furthest reaching forward and reverse paths starting at
// / opposing corners 'overlap'.
func find_middle_snake[O comparable](
	old []O,
	old_range Range,
	new []O,
	new_range Range,
	vf *V,
	vb *V,
	deadline Option[Instant],
) Option[[2]usize] {
	n := old_range.len()
	m := new_range.len()

	// By Lemma 1 in the paper, the optimal edit script length is odd or even as
	// `delta` is odd or even.
	delta := n - m
	odd := delta&1 == 1

	// The initial point at (0, -1)
	vf.set(1, 0)
	// The initial point at (N, M+1)
	vb.set(1, 0)

	// We only need to explore ceil(D/2) + 1
	d_max := max_d(n, m)
	assert(vf.len() >= d_max)
	assert(vb.len() >= d_max)

	for d := range d_max {
		// are we running for too long?
		if deadline_exceeded(deadline) {
			break
		}

		// Forward path
		for k := d; k >= -d; k -= 2 {
			var x usize
			if k == -d || (k != d && vf.get(k-1) < vf.get(k+1)) {
				x = vf.get(k + 1)
			} else {
				x = vf.get(k-1) + 1
			}
			y := x - k

			// The coordinate of the start of a snake
			x0, y0 := x, y
			//  While these sequences are identical, keep moving through the
			//  graph with no cost
			if x < old_range.len() && y < new_range.len() {
				advance := common_prefix_len(
					old,
					Range{old_range.start + x, old_range.end},
					new,
					Range{new_range.start + y, new_range.end},
				)
				x += advance
			}

			// This is the new best x value
			vf.set(k, x)

			// Only check for connections from the forward search when N - M is
			// odd and when there is a reciprocal k line coming from the other
			// direction.
			if odd && (k-delta).abs() <= (d-1) {
				// TODO optimize this so we don't have to compare against n
				if vf.get(k)+vb.get(-(k-delta)) >= n {
					// Return the snake
					return Some([2]usize{x0 + old_range.start, y0 + new_range.start})
				}
			}
		}

		// Backward path
		for k := d; k >= -d; k -= 2 {
			var x usize
			if k == -d || (k != d && vb.get(k-1) < vb.get(k+1)) {
				x = vb.get(k + 1)
			} else {
				x = vb.get(k-1) + 1
			}
			y := x - k

			// The coordinate of the start of a snake
			if x < n && y < m {
				advance := common_suffix_len(
					old,
					Range{old_range.start, old_range.start + n - x},
					new,
					Range{new_range.start, new_range.start + m - y},
				)
				x += advance
				y += advance
			}

			// This is the new best x value
			vb.set(k, x)

			if !odd && (k-delta).abs() <= d {
				// TODO optimize this so we don't have to compare against n
				if vb.get(k)+vf.get(-(k-delta)) >= n {
					// Return the snake
					return Some([2]usize{n - x + old_range.start, m - y + new_range.start})
				}
			}
		}

		// TODO: Maybe there's an opportunity to optimize and bail early?
	}

	// deadline reached
	return None[[2]usize]()
}

func conquer[O comparable, D DiffHook](
	d D,
	old []O,
	old_range Range,
	new []O,
	new_range Range,
	vf *V,
	vb *V,
	deadline Option[Instant],
) Result[Unit] {
	// Check for common prefix
	common_prefix_len := common_prefix_len(old, old_range, new, new_range)
	if common_prefix_len > 0 {
		if tmp := d.equal(old_range.start, new_range.start, common_prefix_len); !tmp.Ok {
			return tmp
		}
	}
	old_range.start += common_prefix_len
	new_range.start += common_prefix_len

	// Check for common suffix
	common_suffix_len := common_suffix_len(old, old_range, new, new_range)
	common_suffix := [2]usize{
		old_range.end - common_suffix_len,
		new_range.end - common_suffix_len,
	}
	old_range.end -= common_suffix_len
	new_range.end -= common_suffix_len

	if is_empty_range(old_range) && is_empty_range(new_range) {
		// Do nothing
	} else if is_empty_range(new_range) {
		if tmp := d.delete(old_range.start, old_range.len(), new_range.start); !tmp.Ok {
			return tmp
		}
	} else if is_empty_range(old_range) {
		if tmp := d.insert(old_range.start, new_range.start, new_range.len()); !tmp.Ok {
			return tmp
		}
	} else if xy_start, ok := find_middle_snake(
		old,
		old_range,
		new,
		new_range,
		vf,
		vb,
		deadline,
	).Unpack(); ok {
		x_start, y_start := xy_start[0], xy_start[1]
		old_a, old_b := split_at(old_range, x_start)
		new_a, new_b := split_at(new_range, y_start)
		if tmp := conquer(d, old, old_a, new, new_a, vf, vb, deadline); !tmp.Ok {
			return tmp
		}
		if tmp := conquer(d, old, old_b, new, new_b, vf, vb, deadline); !tmp.Ok {
			return tmp
		}
	} else {
		if tmp := d.delete(
			old_range.start,
			old_range.end-old_range.start,
			new_range.start,
		); !tmp.Ok {
			return tmp
		}
		if tmp := d.insert(
			old_range.start,
			new_range.start,
			new_range.end-new_range.start,
		); !tmp.Ok {
			return tmp
		}
	}

	if common_suffix_len > 0 {
		return d.equal(common_suffix[0], common_suffix[1], common_suffix_len)
	}

	return Ok(Unit{})
}

// #[test]
// fn test_find_middle_snake() {
//     let a = &b"ABCABBA"[..];
//     let b = &b"CBABAC"[..];
//     let max_d = max_d(a.len(), b.len());
//     let mut vf = V::new(max_d);
//     let mut vb = V::new(max_d);
//     let (x_start, y_start) =
//         find_middle_snake(a, 0..a.len(), b, 0..b.len(), &mut vf, &mut vb, None).unwrap();
//     assert_eq!(x_start, 4);
//     assert_eq!(y_start, 1);
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
// fn test_deadline_reached() {
//     use std::ops::Index;
//     use std::time::Duration;

//     let a = (0..100).collect::<Vec<_>>();
//     let mut b = (0..100).collect::<Vec<_>>();
//     b[10] = 99;
//     b[50] = 99;
//     b[25] = 99;

//     struct SlowIndex<'a>(&'a [usize]);

//     impl Index<usize> for SlowIndex<'_> {
//         type Output = usize;

//         fn index(&self, index: usize) -> &Self::Output {
//             std::thread::sleep(Duration::from_millis(1));
//             &self.0[index]
//         }
//     }

//     let slow_a = SlowIndex(&a);
//     let slow_b = SlowIndex(&b);

//     // don't give it enough time to do anything interesting
//     let mut d = crate::algorithms::Replace::new(crate::algorithms::Capture::new());
//     diff_deadline(
//         &mut d,
//         &slow_a,
//         0..a.len(),
//         &slow_b,
//         0..b.len(),
//         Some(Instant::now() + Duration::from_millis(50)),
//     )
//     .unwrap();
//     insta::assert_debug_snapshot!(d.into_inner().ops());
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
