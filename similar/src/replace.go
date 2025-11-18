package src

// A [`DiffHook`] that combines deletions and insertions to give blocks
// of maximal length, and replacements when appropriate.
//
// It will replace [`DiffHook::insert`] and [`DiffHook::delete`] events when
// possible with [`DiffHook::replace`] events.  Note that even though the
// text processing in the crate does not use replace events and always resolves
// then back to delete and insert, it's useful to always use the replacer to
// ensure a consistent order of inserts and deletes.  This is why for instance
// the text diffing automatically uses this hook internally.
type Replace[D DiffHook] struct {
	d            D
	del, ins, eq Option[[3]usize]
}

func NewReplace[D DiffHook](d D) *Replace[D] {
	none := None[[3]usize]()
	return &Replace[D]{d, none, none, none}
}

// impl<D: DiffHook> Replace<D> {
//     /// Creates a new replace hook wrapping another hook.
//     pub fn new(d: D) -> Self {
//         Replace {
//             d,
//             del: None,
//             ins: None,
//             eq: None,
//         }
//     }

//     /// Extracts the inner hook.
//     pub fn into_inner(self) -> D {
//         self.d
//     }

func (self *Replace[D]) flush_eq() Result[Unit] {
	if eq, ok := self.eq.Unpack(); ok {
		return self.d.equal(eq[0], eq[1], eq[2])
	}
	return Ok(Unit{})
}

func (self *Replace[D]) flush_del_ins() Result[Unit] {
	if del, ok := self.del.Unpack(); ok {
		del_old_index, del_old_len, del_new_index := del[0], del[1], del[2]
		if self.ins.Valid {
			_, ins_new_index, ins_new_len := self.ins.Value[0], self.ins.Value[1], self.ins.Value[2]
			return self.d.replace(del_old_index, del_old_len, ins_new_index, ins_new_len)
		} else {
			return self.d.delete(del_old_index, del_old_len, del_new_index)
		}
	} else if self.ins.Valid {
		ins_old_index, ins_new_index, ins_new_len := self.ins.Value[0], self.ins.Value[1], self.ins.Value[2]
		return self.d.insert(ins_old_index, ins_new_index, ins_new_len)
	}
	return Ok(Unit{})
}

func (self *Replace[D]) equal(old_index, new_index, len usize) Result[Unit] {
	if tmp := self.flush_del_ins(); !tmp.Ok {
		return tmp
	}

	if self.eq.Valid {
		eq_old_index, eq_new_index, eq_len := self.eq.Value[0], self.eq.Value[1], self.eq.Value[2]
		self.eq =
			Some([3]usize{eq_old_index, eq_new_index, eq_len + len})
	} else {
		self.eq =
			Some([3]usize{old_index, new_index, len})
	}

	return Ok(Unit{})
}

func (self *Replace[D]) delete(old_index, old_len, new_index usize) Result[Unit] {
	if tmp := self.flush_eq(); !tmp.Ok {
		return tmp
	}

	if self.del.Valid {
		del_old_index, del_old_len, del_new_index := self.del.Value[0], self.del.Value[1], self.del.Value[2]
		debug_assert_eq(old_index, del_old_index+del_old_len)
		self.del = Some([3]usize{del_old_index, del_old_len + old_len, del_new_index})
	} else {
		self.del = Some([3]usize{old_index, old_len, new_index})
	}
	return Ok(Unit{})
}

func (self *Replace[D]) insert(old_index, new_index, new_len usize) Result[Unit] {
	if tmp := self.flush_eq(); !tmp.Ok {
		return tmp
	}

	if self.ins.Valid {
		ins_old_index, ins_new_index, ins_new_len := self.ins.Value[0], self.ins.Value[1], self.ins.Value[2]
		debug_assert_eq(ins_new_index+ins_new_len, new_index)
		Some([3]usize{ins_old_index, ins_new_index, new_len + ins_new_len})
	} else {
		Some([3]usize{old_index, new_index, new_len})
	}

	return Ok(Unit{})
}

func (self *Replace[D]) replace(old_index, old_len, new_index, new_len usize) Result[Unit] {
	if tmp := self.flush_eq(); !tmp.Ok {
		return tmp
	}

	return self.d.replace(old_index, old_len, new_index, new_len)
}

func (self *Replace[D]) finish() Result[Unit] {
	if tmp := self.flush_eq(); !tmp.Ok {
		return tmp
	}

	if tmp := self.flush_del_ins(); !tmp.Ok {
		return tmp
	}

	return self.d.finish()
}

// #[test]
// fn test_mayers_replace() {
//     use crate::algorithms::{diff_slices, Algorithm};
//     let a: &[&str] = &[
//         ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n",
//         "a\n",
//         "b\n",
//         "c\n",
//         "================================\n",
//         "d\n",
//         "e\n",
//         "f\n",
//         "<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n",
//     ];
//     let b: &[&str] = &[
//         ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n",
//         "x\n",
//         "b\n",
//         "c\n",
//         "================================\n",
//         "y\n",
//         "e\n",
//         "f\n",
//         "<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n",
//     ];

//     let mut d = Replace::new(crate::algorithms::Capture::new());
//     diff_slices(Algorithm::Myers, &mut d, a, b).unwrap();

//     insta::assert_debug_snapshot!(&d.into_inner().ops(), @r###"
//     [
//         Equal {
//             old_index: 0,
//             new_index: 0,
//             len: 1,
//         },
//         Replace {
//             old_index: 1,
//             old_len: 1,
//             new_index: 1,
//             new_len: 1,
//         },
//         Equal {
//             old_index: 2,
//             new_index: 2,
//             len: 3,
//         },
//         Replace {
//             old_index: 5,
//             old_len: 1,
//             new_index: 5,
//             new_len: 1,
//         },
//         Equal {
//             old_index: 6,
//             new_index: 6,
//             len: 3,
//         },
//     ]
//     "###);
// }

// #[test]
// fn test_replace() {
//     use crate::algorithms::{diff_slices, Algorithm};

//     let a: &[usize] = &[0, 1, 2, 3, 4];
//     let b: &[usize] = &[0, 1, 2, 7, 8, 9];

//     let mut d = Replace::new(crate::algorithms::Capture::new());
//     diff_slices(Algorithm::Myers, &mut d, a, b).unwrap();
//     insta::assert_debug_snapshot!(d.into_inner().ops(), @r###"
//     [
//         Equal {
//             old_index: 0,
//             new_index: 0,
//             len: 3,
//         },
//         Replace {
//             old_index: 3,
//             old_len: 2,
//             new_index: 3,
//             new_len: 3,
//         },
//     ]
//     "###);
// }
