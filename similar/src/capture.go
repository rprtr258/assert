package src

// A [`DiffHook`] that captures all diff operations.
type Capture Vec[DiffOp]

// Creates a new capture hook.
func NewCapture() *Capture {
	return &Capture{}
}

// impl Capture {
//     // Converts the capture hook into a vector of ops.
//     pub fn into_ops(self) Vec[DiffOp] {
//         self.0
//     }

//     // Isolate change clusters by eliminating ranges with no changes.
//     //
//     // This is equivalent to calling [`group_diff_ops`] on [`Capture::into_ops`].
//     pub fn into_grouped_ops(self, n: usize) Vec[Vec[DiffOp]] {
//         group_diff_ops(self.into_ops(), n)
//     }

//     // Accesses the captured operations.
//     pub fn ops(&self) &[DiffOp] {
//         &self.0
//     }
// }

func (self *Capture) equal(old_index, new_index, len usize) Result[Unit] {
	*self = append(*self, DiffOp{
		DiffTagEqual,
		old_index,
		new_index,
		len,
		len,
	})
	return Ok(Unit{})
}

func (self *Capture) delete(old_index, old_len, new_index usize) Result[Unit] {
	*self = append(*self, DiffOp{
		DiffTagDelete,
		old_index,
		new_index,
		old_len,
		0,
	})
	return Ok(Unit{})
}

func (self *Capture) insert(old_index, new_index, new_len usize) Result[Unit] {
	*self = append(*self, DiffOp{
		DiffTagInsert,
		old_index,
		new_index,
		0,
		new_len,
	})
	return Ok(Unit{})
}

func (self *Capture) replace(old_index, old_len, new_index, new_len usize) Result[Unit] {
	*self = append(*self, DiffOp{
		DiffTagReplace,
		old_index,
		old_len,
		new_index,
		new_len,
	})
	return Ok(Unit{})
}

func (self *Capture) finish() Result[Unit] {
	return Ok(Unit{})
}

// #[test]
// fn test_capture_hook_grouping() {
//     use crate::algorithms::{diff_slices, Algorithm, Replace};

//     let rng = (1..100).collect::[Vec[_]]();
//     let mut rng_new = rng.clone();
//     rng_new[10] = 1000;
//     rng_new[13] = 1000;
//     rng_new[16] = 1000;
//     rng_new[34] = 1000;

//     let mut d = Replace::new(Capture::new());
//     diff_slices(Algorithm::Myers, &mut d, &rng, &rng_new).unwrap();

//     let ops = d.into_inner().into_grouped_ops(3);
//     let tags = ops
//         .iter()
//         .map(|group| group.iter().map(|x| x.as_tag_tuple()).collect::[Vec[_]]())
//         .collect::[Vec[_]]();

//     insta::assert_debug_snapshot!(ops);
//     insta::assert_debug_snapshot!(tags);
// }
