package core


type hook struct {
	next *hook
	f interface{}
}

type hooklist struct {
	all *hook
	regged *hook
}

func (l *hooklist) add(f interface{}, unregged bool) {
	h := new(hook)
	h.f = f

	var list **hook
	if unregged {
		list = &l.all
	} else {
		list = &l.regged
	}

	for *list != nil {
		list = &((*list).next)
	}
	*list = h
}

func (l *hooklist) run(f func(interface{}), registered bool) {
	for h := l.all; h != nil; h = h.next {
		f(h.f)
	}

	if !registered { return }
	
	for h := l.regged; h != nil; h = h.next {
		f(h.f)
	}
}
