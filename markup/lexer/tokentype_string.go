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
	_ = x[TokenMono-13]
	_ = x[TokenCodeBlockBegin-14]
	_ = x[TokenCodeBlockEnd-15]
	_ = x[TokenText-16]
	_ = x[TokenEmphasis-17]
	_ = x[TokenStrong-18]
	_ = x[TokenEmphasisStrong-19]
	_ = x[TokenLinkHref-20]
	_ = x[TokenLinkText-21]
	_ = x[TokenAmpSpecial-22]
	_ = x[TokenEscaped-23]
	_ = x[TokenBlockquoteBegin-24]
	_ = x[TokenBlockquoteEnd-25]
	_ = x[TokenEnquoteBegin-26]
	_ = x[TokenEnquoteEnd-27]
	_ = x[TokenImage-28]
	_ = x[TokenImageTitle-29]
	_ = x[TokenImagePath-30]
	_ = x[TokenImageAlt-31]
	_ = x[TokenHorizontalRule-32]
}

const _TokenType_name = "EOFMetaBeginMetaKeyMetaValMetaEndHtmlTagOpenHtmlTagAttrKeyHtmlTagAttrValHtmlTagCloseParagraphBeginParagraphEndSection1Section2MonoCodeBlockBeginCodeBlockEndTextEmphasisStrongEmphasisStrongLinkHrefLinkTextAmpSpecialEscapedBlockquoteBeginBlockquoteEndEnquoteBeginEnquoteEndImageImageTitleImagePathImageAltHorizontalRule"

var _TokenType_index = [...]uint16{0, 3, 12, 19, 26, 33, 44, 58, 72, 84, 98, 110, 118, 126, 130, 144, 156, 160, 168, 174, 188, 196, 204, 214, 221, 236, 249, 261, 271, 276, 286, 295, 303, 317}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
