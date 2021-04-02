package cff

type SubroutineSelector interface {
	GetSubroutines(gid int) (global []pscmdseq, local []pscmdseq)
}

type FontDictionarySelect interface {
	GetFontDictionaryIndex(gid int) int
}

type format0FdSelect struct {
	fds []int
}

func (fd *format0FdSelect) GetFontDictionaryIndex(gid int) int {
	if gid < len(fd.fds) && gid >= 0 {
		return fd.fds[gid]
	}
	return 0
}

type range3 struct {
	first int
	fd    int
}

type format1FdSelect struct {
	ranges   []range3
	sentinel int
}

func (fd *format1FdSelect) GetFontDictionaryIndex(gid int) int {
	for i := 0; i < len(fd.ranges); i++ {
		if fd.ranges[i].first <= gid {
			if i+1 < len(fd.ranges) {
				if fd.ranges[i+1].first > gid {
					return fd.ranges[i].fd
				}
			} else {
				if fd.sentinel > gid {
					return fd.ranges[i].fd
				}
				return -1
			}
		}
	}
	return 0
}

type fontdict struct {
	tld         *TopLevelDictionary
	private     *PrivateDictionary
	subroutines []pscmdseq
}

type selector struct {
	dict     []*fontdict
	fdselect FontDictionarySelect
}

func newselector(d *fontdict) *selector {
	return &selector{
		dict: []*fontdict{d},
	}
}

func newcidselector(d []*fontdict, sel FontDictionarySelect) *selector {
	return &selector{
		dict:     d,
		fdselect: sel,
	}
}

func (s *selector) GetFontDictionary(gid int) *fontdict {
	if s.fdselect == nil {
		return s.dict[0]
	}

	idx := s.fdselect.GetFontDictionaryIndex(gid)
	if idx < 0 || idx >= len(s.dict) {
		return nil
	}
	return s.dict[idx]
}

type subroutineselector struct {
	local    []pscmdseq
	global   []pscmdseq
	selector *selector
}

func newsubselector(g []pscmdseq, l []pscmdseq, sel *selector) SubroutineSelector {
	return &subroutineselector{local: l, global: g, selector: sel}
}

func (s *subroutineselector) GetSubroutines(gid int) (global []pscmdseq, local []pscmdseq) {
	d := s.selector.GetFontDictionary(gid)
	local = d.subroutines
	if local == nil {
		local = s.local
	}
	return s.global, local

	// if !s.cid {
	// 	return s.global, s.local
	// }
	// idx := s.fdselect.GetFontDictionaryIndex(gid)
	// if idx < 0 || idx >= len(s.subroutines) {
	// 	return s.global, s.local
	// }
	// local = s.subroutines[idx]
	// if local == nil {
	// 	local = s.local
	// }
	// return s.global, local
}
