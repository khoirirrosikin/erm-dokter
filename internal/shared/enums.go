package shared

import "strings"

type StatusLanjut string

const (
	StatusLanjutRawatJalan StatusLanjut = "Ralan"
	StatusLanjutRawatInap  StatusLanjut = "Ranap"
	StatusLanjutSemua      StatusLanjut = "Semua"
)

func (s StatusLanjut) IsValid() bool {
	switch s {
	case StatusLanjutRawatJalan, StatusLanjutRawatInap:
		return true
	default:
		return false
	}
}

func ParseStatusLanjut(val string) (StatusLanjut, bool) {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "ralan":
		return StatusLanjutRawatJalan, true
	case "ranap":
		return StatusLanjutRawatInap, true
	default:
		return "", false
	}
}

func ParseStatusLanjutWithSemua(val string) (StatusLanjut, bool) {
	if strings.ToLower(strings.TrimSpace(val)) == "semua" {
		return StatusLanjutSemua, true
	}
	return ParseStatusLanjut(val)
}

type SortOrder string

const (
	SortASC  SortOrder = "ASC"
	SortDESC SortOrder = "DESC"
)

func (s SortOrder) IsValid() bool {
	switch s {
	case SortASC, SortDESC:
		return true
	default:
		return false
	}
}

type KategoriLab string

const (
	KategoriLabPK KategoriLab = "PK"
	KategoriLabPA KategoriLab = "PA"
	KategoriLabMB KategoriLab = "MB"
)

func (k KategoriLab) IsValid() bool {
	switch k {
	case KategoriLabPK, KategoriLabPA, KategoriLabMB:
		return true
	default:
		return false
	}
}
