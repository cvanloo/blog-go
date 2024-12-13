// Code generated by "stringer -type TokenType -trimprefix Token"; DO NOT EDIT.

package lexer

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TokenEOF-0]
	_ = x[TokenMetaBegin-1]
	_ = x[TokenMetaKey-2]
	_ = x[TokenMetaVal-3]
	_ = x[TokenMetaEnd-4]
	_ = x[TokenHtmlTagOpen-5]
	_ = x[TokenHtmlTagAttrKey-6]
	_ = x[TokenHtmlTagAttrVal-7]
	_ = x[TokenHtmlTagClose-8]
	_ = x[TokenParagraphBegin-9]
	_ = x[TokenParagraphEnd-10]
	_ = x[TokenSection1-11]
	_ = x[TokenSection2-12]
	_ = x[TokenCodeBlockBegin-13]
	_ = x[TokenCodeBlockEnd-14]
	_ = x[TokenText-15]
}

const _TokenType_name = "EOFMetaBeginMetaKeyMetaValMetaEndHtmlTagOpenHtmlTagAttrKeyHtmlTagAttrValHtmlTagCloseParagraphBeginParagraphEndSection1Section2CodeBlockBeginCodeBlockEndText"

var _TokenType_index = [...]uint8{0, 3, 12, 19, 26, 33, 44, 58, 72, 84, 98, 110, 118, 126, 140, 152, 156}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
