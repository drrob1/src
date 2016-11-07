package poetry;

import (
	"testing"
)

func TestNumLines(t *testing.T) {
	p := Poem{}; // This is part of same poetry package, so poetry.Poem here would be an error.
	if p.GetNumLines() != 0 {
	  t.Fatalf(" Empty poem is not empty: unexpected stanza count %d",p.GetNumStanzas());
	}

	p = Poem{{"the quick","brown fox","jumps over","the lazy dog"}};
	if p.GetNumLines() != 4 {
	  t.Fatalf(" unexpected stanza count %d",p.GetNumStanzas());
	}
}

func TestStats(t *testing.T) {
	p := Poem{};
	v,c,u := p.Stats();
	if (v != 0) || (c != 0) || (u != 0) {
	  t.Fatalf(" v,c and u should be zero.  v=%d, c=%d, u=%d",v,c,u);
	}
	p = Poem{{"Hello"}};
	v,c,u = p.Stats();
	if (v != 2) || (c != 3) || (u != 0) {
	  t.Fatalf(" v and c not correct for Hello.  v=%d, c=%d, u=%d",v,c,u);
	}
	p = Poem{{"Hello, World!"}};
	v,c,u = p.Stats();
	if (v != 3) || (c != 7) || (u != 3) {
	  t.Fatalf(" v and c not correct for Hello, World!.  v=%d, c=%d, u=%d",v,c,u);
	}
}

func TestNumWords(t *testing.T) {
	p := Poem{};
	if p.GetNumWords() != 0 {
	  t.Fatalf(" wrong # of words in an empty poem.  It came back %d \n",p.GetNumWords());
	}

	p = Poem{{"Hello, World!"}};
	if p.GetNumWords() != 2 {
	  t.Fatalf(" wrong # of words in poem.  It should be 2 but came back %d \n",p.GetNumWords());
	}

	p = Poem{{"The Quick Brown Fox Jumped Over The Lazy Dog"}};
	if p.GetNumWords() != 9 {
	  t.Fatalf(" wrong # of words in poem.  It should be 9 came back %d \n",p.GetNumWords());
	}
}
