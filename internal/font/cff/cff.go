package cff

import "fmt"

type strtable []string

func (si strtable) GetName(index int) string {
	if index >= 0 && index <= 390 {
		return getStandardStringName(index)
	}

	strIndexIdx := index - 391
	if strIndexIdx >= 0 && strIndexIdx <= len(si) {
		return si[strIndexIdx]
	}

	return fmt.Sprintf("SID%d", index)
}

type FDSelect interface {
	GetFontDictionaryIndex(gid int) int
}

type fmtzerofdselect struct {
	ros RegistryOrderingSupplement
	fds []int
}

func (fd *fmtzerofdselect) GetFontDictionaryIndex(gid int) int {
	if gid < len(fd.fds) && gid >= 0 {
		return fd.fds[gid]
	}
	return 0
}

type range3 struct {
	first int
	fd    int
}

type fmtonefdselect struct {
	ros      RegistryOrderingSupplement
	ranges   []range3
	sentinel int
}

func (fd *fmtonefdselect) GetFontDictionaryIndex(gid int) int {
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

type subroutineselector struct {
	global      table
	local       table
	cid         bool
	fdselect    FDSelect
	subroutines []table
}

func (s *subroutineselector) subroutine(gid int) (table, table) {
	if !s.cid {
		return s.global, s.local
	}
	idx := s.fdselect.GetFontDictionaryIndex(gid)
	if idx < 0 || idx >= len(s.subroutines) {
		return s.global, s.local
	}
	local := s.subroutines[idx]
	if local == nil {
		local = s.local
	}
	return s.global, local
}
