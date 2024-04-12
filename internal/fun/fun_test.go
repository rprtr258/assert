func TestFlatMap(t *testing.T) {
 seq := FromMany(1, 2, 3)
 flatMappedSeq := FlatMap(seq, func(n int) Seq[int] {
  return FromMany(n, n*n)
 })

 var result []int
 flatMappedSeq(func(n int) bool {
  result = append(result, n)
  return true
 })

 expected := []int{1, 1, 2, 4, 3, 9}
 for i, v := range expected {
  if result[i] != v {
   t.Errorf("FlatMap element %d = %d; want %d", i, result[i], v)
  }
 }
}


  seq := FromMany(1, 2, 3)
  mappedSeq := Map(seq, func(n int) int {
    return n * n
  })

  var result []int
  mappedSeq(func(n int) bool {
    result = append(result, n)
    return true
  })

  expected := []int{1, 4, 9}
  for i, v := range expected {
    if result[i] != v {
      t.Errorf("Map element %d = %d; want %d", i, result[i], v)
    }
  }
}


	dict := map[string]int{"one": 1, "two": 2}
	var keys []string
	FromDictKeys(dict)(func(k string) bool {
		keys = append(keys, k)
		return true
	})

	if len(keys) != 2 || !(keys[0] == "one" && keys[1] == "two" || keys[0] == "two" && keys[1] == "one") {
		t.Errorf("FromDictKeys keys = %v; want ['one', 'two'] or ['two', 'one']", keys)
	}
}

	var result []int
	FromRange(1, 4)(func(n int) bool {
		result = append(result, n)
		return true
	})

	expected := []int{1, 2, 3}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("FromRange element %d = %d; want %d", i, result[i], v)
		}
	}
}func TestFromMany(t *testing.T) {
	var result []int
	FromMany(1, 2, 3)(func(n int) bool {
		result = append(result, n)
		return true
	})

	expected := []int{1, 2, 3}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("FromMany element %d = %d; want %d", i, result[i], v)
		}
	}
}

