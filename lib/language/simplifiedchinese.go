package language

import (
	"golang.org/x/text/encoding/simplifiedchinese"
)

type Charset int

const (
	UNKNOWN Charset = iota
	UTF8
	GB18030
	GBK
	HZGB2312
)

func (c Charset) String() string {
	switch c {
	case UTF8:
		return "UTF-8"
	case GB18030:
		return "GB18030"
	case GBK:
		return "GBK"
	case HZGB2312:
		return "HZGB2312"
	default:
		return "UTF-8"
	}
}

func BytesToStringQuick(charset Charset, buf []byte) string {
	var str string
	switch charset {
	case GB18030:
		tmp, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(buf)
		str = string(tmp)
	case GBK:
		tmp, _ := simplifiedchinese.GBK.NewDecoder().Bytes(buf)
		str = string(tmp)
	case HZGB2312:
		tmp, _ := simplifiedchinese.HZGB2312.NewDecoder().Bytes(buf)
		str = string(tmp)
	case UTF8:
		fallthrough
	default:
		str = string(buf)
	}
	return str
}
