// Code generated by "stringer -type ParseState"; DO NOT EDIT.

package parser

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ParsingStart-0]
	_ = x[ParsingMeta-1]
	_ = x[ParsingMetaVal-2]
	_ = x[ParsingDocument-3]
	_ = x[ParsingSection1-4]
	_ = x[ParsingSection1Content-5]
	_ = x[ParsingSection2-6]
	_ = x[ParsingSection2Content-7]
	_ = x[ParsingHtmlTag-8]
	_ = x[ParsingHtmlTagContent-9]
	_ = x[ParsingParagraph-10]
	_ = x[ParsingCodeBlock-11]
	_ = x[ParsingImage-12]
	_ = x[ParsingImageTitle-13]
	_ = x[ParsingImageAlt-14]
	_ = x[ParsingBlockquote-15]
	_ = x[ParsingBlockquoteAttrAuthor-16]
	_ = x[ParsingBlockquoteAttrSource-17]
	_ = x[ParsingEnquoteSingle-18]
	_ = x[ParsingEnquoteDouble-19]
	_ = x[ParsingEnquoteAngled-20]
	_ = x[ParsingEmphasis-21]
	_ = x[ParsingStrong-22]
	_ = x[ParsingEmphasisStrong-23]
	_ = x[ParsingLink-24]
}

const _ParseState_name = "ParsingStartParsingMetaParsingMetaValParsingDocumentParsingSection1ParsingSection1ContentParsingSection2ParsingSection2ContentParsingHtmlTagParsingHtmlTagContentParsingParagraphParsingCodeBlockParsingImageParsingImageTitleParsingImageAltParsingBlockquoteParsingBlockquoteAttrAuthorParsingBlockquoteAttrSourceParsingEnquoteSingleParsingEnquoteDoubleParsingEnquoteAngledParsingEmphasisParsingStrongParsingEmphasisStrongParsingLink"

var _ParseState_index = [...]uint16{0, 12, 23, 37, 52, 67, 89, 104, 126, 140, 161, 177, 193, 205, 222, 237, 254, 281, 308, 328, 348, 368, 383, 396, 417, 428}

func (i ParseState) String() string {
	if i < 0 || i >= ParseState(len(_ParseState_index)-1) {
		return "ParseState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ParseState_name[_ParseState_index[i]:_ParseState_index[i+1]]
}
