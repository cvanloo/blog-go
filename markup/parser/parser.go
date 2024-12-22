package parser

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"strconv"
	"log"

	//"github.com/kr/pretty"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/gen"
)

type (
	LexResult interface {
		Tokens() func(func(lexer.Token) bool)
	}
	ParserError struct {
		State ParseState
		Token lexer.Token
		Inner error
	}
)

// @todo: make parser a type like with the lexer?
func newError(lexeme lexer.Token, state ParseState, inner error) error {
	if inner == nil {
		return nil
	}
	return ParserError{
		State: state,
		Token: lexeme,
		Inner: inner,
	}
}

func (err ParserError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", err.State, err.Token, err.Inner)
}

type (
	Level struct {
		ReturnToState ParseState
		Strings Stack[string]
		Content Stack[gen.Renderable]
		TextValues Stack[gen.StringRenderable]
	}
	Levels struct {
		levels []*Level
	}
)

func (lv *Level) PushString(v string) {
	lv.Strings = lv.Strings.Push(v)
}

func (lv *Level) PopString() (s string) {
	lv.Strings, s = lv.Strings.Pop()
	return s
}

func (lv *Level) PushContent(v gen.Renderable) {
	lv.Content = lv.Content.Push(v)
}

func (lv *Level) PushText(v gen.StringRenderable) {
	lv.TextValues = lv.TextValues.Push(v)
}

func (lv *Level) EmptyText() {
	lv.TextValues = lv.TextValues.Empty()
}

func (ls *Levels) Push(l *Level) {
	ls.levels = append(ls.levels, l)
}

func (ls *Levels) Top() *Level {
	l := len(ls.levels)
	return ls.levels[l-1]
}

func (ls *Levels) Dig() *Level {
	l := len(ls.levels)
	return ls.levels[l-2]
}

func (ls *Levels) Pop() *Level {
	l := len(ls.levels)
	top := ls.levels[l-1]
	ls.levels = ls.levels[:l-1]
	return top
}

func (ls *Levels) Len() int {
	return len(ls.levels)
}

//go:generate stringer -type ParseState
type ParseState int

const (
	ParsingStart ParseState = iota
	ParsingDocument
	ParsingMeta
	ParsingMetaVal
	ParsingAttributeList
	ParsingSection1
	ParsingSection1AfterAttributeList
	ParsingSection1Content
	ParsingSection2
	ParsingSection2AfterAttributeList
	ParsingSection2Content
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrInvalidMetaKey = errors.New("invalid meta key")
	ErrSectionMissingHeading = errors.New("section must have a heading")
)

func Parse(lx LexResult) (blog gen.Blog, err error) {
	state := ParsingStart
	levels := Levels{}
	levels.Push(&Level{ReturnToState: ParsingStart})
	var (
		currentSection1, currentSection2 gen.Section
		currentAttributes gen.Attributes
		currentHtmlTag = HtmlTag{Args: map[string]string{}}
	)
	for lexeme := range lx.Tokens() {
		level := levels.Top()
		switch state {
		default:
			panic(fmt.Errorf("parser state not implemented: %s", state))
		case ParsingStart:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenMetaBegin:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingMeta
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingHtmlTag
			case lexer.TokenSection1Begin:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingSection1
			case lexer.TokenDefinitionTerm:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingTermDefinition
			case lexer.TokenLinkDef:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingLinkDefinition
			case lexer.TokenSidenoteDef:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingSidenoteDefinition
			case lexer.TokenEOF:
				levels.Pop()
				Assert(levels.Len() == 0, "not all levels popped")
			}
		case ParsingDocument:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingHtmlTag
			case lexer.TokenSection1Begin:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingSection1
			case lexer.TokenDefinitionTerm:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingTermDefinition
			case lexer.TokenLinkDef:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingLinkDefinition
			case lexer.TokenSidenoteDef:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingSidenoteDefinition
			case lexer.TokenEOF:
				levels.Pop()
				Assert(levels.Len() == 0, "not all levels popped")
			}
		case ParsingMeta:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenMetaKey:
				if !checkMetaKey(lexeme.Text) {
					err = errors.Join(err, newError, ErrInvalidMetaKey))
				}
				level.PushString(lexeme.Text)
				state = ParsingMetaVal
			case lexer.TokenMetaEnd:
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingMetaVal:
			switch lexeme.Type {
			default:
				if isTextContent(lexeme.Type) {
					level.PushText(newTextContent(lexeme))
				} else {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenMetaKey:
				// finish current key
				key := level.PopString()
				setMetaKeyValuePair(&blog, metaKey, gen.StringOnlyContent(level.TextValues))
				level.EmptyText()
				// start next key
				if !checkMetaKey(lexeme.Text) {
					err = errors.Join(err, newError, ErrInvalidMetaKey))
				}
				level.PushString(lexeme.Text)
			case lexer.TokenMetaEnd:
				// finish last key
				key := level.PopString()
				setMetaKeyValuePair(&blog, metaKey, gen.StringOnlyContent(level.TextValues))
				level.EmptyText()
				// finish level
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingAttributeList:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenAttributeListID:
				currentAttributes.ID = lexeme.Text
				state = ParsingAttributeListAfterID
			case lexer.TokenAttributeListKey:
				level.PushString(lexeme.Text)
				state = ParsingAttributeListVal
			case lexer.TokenAttributeListEnd:
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingAttributeListAfterID:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenAttributeListKey:
				level.PushString(lexeme.Text)
				state = ParsingAttributeListVal
			case lexer.TokenAttributeListEnd:
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingAttributeListVal:
			switch lexeme.Type {
			default:
				if isTextContent(lexeme.Type) {
					level.PushText(newTextContent(lexeme))
				} else {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenAttributeListKey:
				// finish previous key
				key := level.PopString()
				val := gen.StringOnlyContent(level.TextValues)
				currentAttributes.SetAttr(key, val.Text())
				// start next key
				level.PushString(lexeme.Text)
			case lexer.TokenAttributeListEnd:
				// finish last key
				key := level.PopString()
				val := gen.StringOnlyContent(level.TextValues)
				currentAttributes.SetAttr(key, val.Text())
				// finish level
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingSection1:
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenAttributeListBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1AfterAttributeList})
				state = ParsingAttributeList
			case lexeme.Type == lexer.TokenSection1Content:
				if len(level.TextValues) == 0 {
					err = errors.Join(err, newError(lexeme, state, ErrSectionMissingHeading))
				}
				currentSection1 = gen.Section{
					Level: 1,
					Heading: gen.StringOnlyContent(level.TextValues),
				}
				level.EmptyText()
				state = ParsingSection1Content
			}
		case ParsingSection1AfterAttributeList:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenSection1Content:
				if len(level.TextValues) == 0 {
					err = errors.Join(err, newError(lexeme, state, ErrSectionMissingHeading))
				}
				currentSection1 = gen.Section{
					Attributes: currentAttributes,
					Level: 1,
					Heading: gen.StringOnlyContent(level.TextValues),
				}
				level.EmptyText()
				currentAttributes = gen.Attributes{}
				state = ParsingSection1Content
			}
		case ParsingSection1Content:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenDefinitionTerm:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingTermDefinition
			case lexer.TokenHorizontalRule:
				level.PushContent(gen.HorizontalRule{})
			case lexer.TokenCodeBlockBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingCodeBlock
			case lexer.TokenSection2Begin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingSection2
			case lexer.TokenImageBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingImage
			case lexer.TokenBlockquoteBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingBlockquote
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				currentHTMLElement.Name = lexeme.Text
				state = ParsingHtmlElement
			case lexer.TokenLinkDef:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingLinkDefinition
			case lexer.TokenSidenoteDef:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingSidenoteDefinition
			case lexer.TokenParagraphBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingParagraph
			case lexer.TokenSection1End:
				currentSection1.Content = level.Content
				blog.Sections = append(blog.Sections, currentSection1)
				currentSection1 = gen.Section{}
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingSection2:
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenAttributeListBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2AfterAttributeList})
				state = ParsingAttributeList
			case lexeme.Type == lexer.TokenSection2Content:
				if len(level.TextValues) == 0 {
					err = errors.Join(err, newError(lexeme, state, ErrSectionMissingHeading))
				}
				currentSection2 = gen.Section{
					Level: 2,
					Heading: gen.StringOnlyContent(level.TextValues),
				}
				level.EmptyText()
				state = ParsingSection2Content
			}
		case ParsingSection2AfterAttributeList:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenSection2Content:
				if len(level.TextValues) == 0 {
					err = errors.Join(err, newError(lexeme, state, ErrSectionMissingHeading))
				}
				currentSection2 = gen.Section{
					Attributes: currentAttributes,
					Level: 2,
					Heading: gen.StringOnlyContent(level.TextValues),
				}
				level.EmptyText()
				currentAttributes = gen.Attributes{}
				state = ParsingSection1Content
			}
		case ParsingSection2Content:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenDefinitionTerm:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingTermDefinition
			case lexer.TokenHorizontalRule:
				level.PushContent(gen.HorizontalRule{})
			case lexer.TokenCodeBlockBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingCodeBlock
			case lexer.TokenImageBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingImage
			case lexer.TokenBlockquoteBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingBlockquote
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				currentHTMLElement.Name = lexeme.Text
				state = ParsingHtmlElement
			case lexer.TokenLinkDef:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingLinkDefinition
			case lexer.TokenSidenoteDef:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingSidenoteDefinition
			case lexer.TokenParagraphBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingParagraph
			case lexer.TokenSection2End:
				currentSection2.Content = level.Content
				levels.Pop()
				parent := levels.Top()
				parent.PushContent(currentSection2)
				currentSection2 = gen.Section{}
				state = level.ReturnToState
			}
		case ParsingParagraph:
		}
	}
	return
}

func Parse(lx LexResult) (blog gen.Blog, err error) {
	state := ParsingStart
	levels := Levels{}
	var (
		currentBlockquote = gen.Blockquote{}
		currentImage = gen.Image{}
		currentHtmlTag = HtmlTag{Args: map[string]string{}}
		//currentSection1 = gen.Section{Level: 1}
		//currentSection2 = gen.Section{Level: 2}
	)
	for lexeme := range lx.Tokens() {
		//log.Printf("[%s/%s] %# v", state, lexeme, pretty.Formatter(levels))
		switch state {
		case ParsingParagraph:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token, expected one of ParagraphEnd, or text content")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenParagraphEnd:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushContent(gen.Paragraph{
					Content: gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			case lexeme.Type == lexer.TokenEmphasisBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEmphasis
			case lexeme.Type == lexer.TokenStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEnquoteSingleBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEnquoteSingle
			case lexeme.Type == lexer.TokenEnquoteDoubleBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEnquoteDouble
			case lexeme.Type == lexer.TokenEnquoteAngledBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEnquoteAngled
			//case lexeme.Type == lexer.TokenLink:
			//	levels.Push(&Level{ReturnToState: ParsingParagraph})
			//	state = ParsingLink
			case lexeme.Type == lexer.TokenHtmlTagOpen:
				// @todo: this is a tad bit tricky (StringRenderable is part of the paragraph, Renderable ends the paragraph, and comes after it)
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				currentHtmlTag.Name = lexeme.Text
				state = ParsingHtmlTag
			}
		case ParsingHtmlTag:
			level := levels.Top()
			switch lexeme.Type {
			default:
			case lexer.TokenHtmlTagAttrKey:
				level.PushString(lexeme.Text)
			case lexer.TokenHtmlTagAttrVal:
				attrKey := level.PopString()
				currentHtmlTag.Args[attrKey] = lexeme.Text
			case lexer.TokenHtmlTagContent:
				Assert(len(level.Strings) == 0, "key/value pair mismatch")
				state = ParsingHtmlTagContent
			case lexer.TokenHtmlTagClose:
				_ = levels.Pop()
				content, evalErr := evaluateHtmlTag(&blog, currentHtmlTag)
				err = errors.Join(err, evalErr)
				if content != nil {
					if levels.Len() == 0 {
						err = errors.Join(err, newError(lexeme, state, errors.New("all content must be contained within a section")))
					} else {
						paren := levels.Top()
						switch c := content.(type) {
						case gen.StringRenderable:
							paren.PushText(c)
						case gen.Renderable:
							paren.PushContent(c)
						}
					}
				}
				currentHtmlTag = HtmlTag{Args: map[string]string{}}
				state = level.ReturnToState
			}
		case ParsingHtmlTagContent:
			level := levels.Top()
			switch {
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
				fallthrough
			case lexeme.Type == lexer.TokenText:
				level.PushString(lexeme.Text)
			case lexeme.Type == lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingHtmlTagContent})
				state = ParsingHtmlTag
			case lexeme.Type == lexer.TokenHtmlTagClose:
				_ = levels.Pop()
				currentHtmlTag.Strings = level.Strings
				currentHtmlTag.Text = level.TextValues
				content, evalErr := evaluateHtmlTag(&blog, currentHtmlTag)
				err = errors.Join(err, newError(lexeme, state, evalErr))
				if content != nil {
					if levels.Len() == 0 {
						err = errors.Join(err, newError(lexeme, state, errors.New("all content must be contained within a section")))
					} else {
						paren := levels.Top()
						switch c := content.(type) {
						case gen.StringRenderable:
							paren.PushText(c)
						case gen.Renderable:
							paren.PushContent(c)
						}
					}
				}
				currentHtmlTag = HtmlTag{Args: map[string]string{}}
				state = level.ReturnToState
			}
		case ParsingCodeBlock:
			level := levels.Top()
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case lexer.TokenCodeBlockLang: // @todo
			case lexer.TokenCodeBlockLineFirst:
			case lexer.TokenCodeBlockLineLast:
			case lexer.TokenText:
				level.PushString(lexeme.Text)
			case lexer.TokenCodeBlockEnd:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushContent(gen.CodeBlock{
					Lines: level.Strings,
				})
				state = level.ReturnToState
			}
		case ParsingImage:
			level := levels.Top()
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case lexer.TokenImageTitle:
				levels.Push(&Level{ReturnToState: ParsingImage})
				state = ParsingImageTitle
			case lexer.TokenImageAlt:
				levels.Push(&Level{ReturnToState: ParsingImage})
				state = ParsingImageAlt
			case lexer.TokenImagePath:
				currentImage.Name = lexeme.Text
			case lexer.TokenImageEnd:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushContent(currentImage)
				currentImage = gen.Image{}
				state = level.ReturnToState
			}
		case ParsingImageTitle:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenImageAttrEnd:
				_ = levels.Pop()
				currentImage.Title = gen.StringOnlyContent(level.TextValues)
				Assert(level.ReturnToState == ParsingImage, "confused parser state")
				state = level.ReturnToState
			}
		case ParsingImageAlt:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenImageAttrEnd:
				_ = levels.Pop()
				currentImage.Alt = gen.StringOnlyContent(level.TextValues)
				Assert(level.ReturnToState == ParsingImage, "confused parser state")
				state = level.ReturnToState
			}
		case ParsingBlockquote:
			level := levels.Top()
			// @todo: also allow links inside blockquote
			switch lexeme.Type {
			default:
				if isTextContent(lexeme.Type) {
					level.PushText(newTextContent(lexeme))
				} else {
					err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
				}
			case lexer.TokenBlockquoteAttrAuthor:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingBlockquoteAttrAuthor
			case lexer.TokenBlockquoteAttrSource:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingBlockquoteAttrSource
			case lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingEmphasis
			case lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingStrong
			case lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingEmphasisStrong
			case lexer.TokenBlockquoteEnd:
				currentBlockquote.QuoteText = gen.StringOnlyContent(level.TextValues)
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushContent(currentBlockquote)
				currentBlockquote = gen.Blockquote{}
				state = level.ReturnToState
			}
		case ParsingBlockquoteAttrAuthor:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenBlockquoteAttrEnd:
				_ = levels.Pop()
				currentBlockquote.Author = gen.StringOnlyContent(level.TextValues)
				Assert(level.ReturnToState == ParsingBlockquote, "confused parser state")
				state = level.ReturnToState
			}
		case ParsingBlockquoteAttrSource:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenBlockquoteAttrEnd:
				_ = levels.Pop()
				currentBlockquote.Source = gen.StringOnlyContent(level.TextValues)
				Assert(level.ReturnToState == ParsingBlockquote, "confused parser state")
				state = level.ReturnToState
			}
		case ParsingEnquoteSingle:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("enquote can only contain text content")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingEnquoteSingle})
				state = ParsingEmphasis
			case lexeme.Type == lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingEnquoteSingle})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingEnquoteSingle})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEnquoteEnd:
				level := levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.EnquoteSingle{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingEnquoteDouble:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("enquote can only contain text content")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingEmphasis
			case lexeme.Type == lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEnquoteEnd:
				level := levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.EnquoteDouble{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingEnquoteAngled:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("enquote can only contain text content")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingEmphasis
			case lexeme.Type == lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEnquoteEnd:
				level := levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.EnquoteAngled{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingEmphasis:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEmphasis:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.Emphasis{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingStrong:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenStrong:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.Strong{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingEmphasisStrong:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.EmphasisStrong{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingLink:
			level := levels.Top()
			switch {
			default:
				if isTextContent(lexeme.Type) {
					level.PushText(newTextContent(lexeme))
				} else {
					err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
				}
			case lexeme.Type == lexer.TokenLinkHref:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.Link{
					Href: lexeme.Text,
					Name: gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		}
	}
	return
}

func isTextContent(tokenType lexer.TokenType) bool {
	switch tokenType {
	default:
		return false
	case lexer.TokenMono, lexer.TokenText, lexer.TokenAmpSpecial:
		// @todo: what else is text content?
		return true
	}
	panic("unreachable")
}

func newTextContent(lexeme lexer.Token) gen.StringRenderable {
	Assert(isTextContent(lexeme.Type), fmt.Sprintf("cannot make text content out of %s", lexeme.Type))
	switch lexeme.Type {
	case lexer.TokenMono:
		return gen.Mono(lexeme.Text)
	case lexer.TokenText:
		return gen.Text(lexeme.Text)
	case lexer.TokenAmpSpecial:
		switch lexeme.Text {
		default:
			panic("programmer error: lexer and parser out of sync about what constitutes a &<...>; special character")
		case "~", "\u00A0":
			return gen.NoBreakSpace
		case "---", "&mdash;":
			return gen.EMDash
		case "&ldquo;":
			return gen.LeftDoubleQuote
		case "&rdquo;":
			return gen.RightDoubleQuote
		case "...", "…":
			return gen.Ellipsis
		case "&prime;":
			return gen.Prime
		case "&Prime;":
			return gen.DoublePrime
		case "&tprime;":
			return gen.TripplePrime
		case "&qprime;":
			return gen.QuadruplePrime
		case "&bprime;":
			return gen.ReversedPrime
		}
	}
	panic("unreachable")
}

func checkMetaKey(key string) bool {
	recognizedKeys := map[string]struct{}{
		"author": {},
		"email": {},
		"tags": {},
		"template": {},
		"title": {},
		"alt-title": {},
		"url-path": {},
		"rel-me": {}, // @todo: this should actually be an array
		"fedi-creator": {},
		"lang": {},
		"published": {},
		"revised": {},
		"est-reading": {},
		"series": {}, // @todo: only parse `series` name and then determine the correct order (with prev and next) depending on all other parsed articles that have the same series name and their published dates.
		"enable-revision-warning": {},
	}
	_, ok := recognizedKeys[key]
	return ok
}

func setMetaKeyValuePair(blog *gen.Blog, key string, value gen.StringRenderable) (err error) {
	switch key {
	default:
		// do nothing, error already reported
	case "author":
		blog.Author.Name = value
	case "email":
		blog.Author.Email = value
	case "tags":
		for _, tag := range strings.Split(value./*@todo: Clean*/Text(), " ") {
			blog.Tags = append(blog.Tags, gen.Tag(tag))
		}
	case "title":
		blog.Title = value // @todo: automatically apply proper English Title Casing
	case "alt-title":
		blog.AltTitle = value
	case "url-path":
		blog.UrlPath = value./*Clean*/Text()
	case "rel-me":
		blog.Author.RelMe = value
	case "fedi-creator":
		blog.Author.FediCreator = value
	case "lang":
		blog.Lang = value./*Clean*/Text()
	case "published":
		blog.Published.Published, err = time.Parse("2006-01-02", value./*Clean*/Text())
		if err != nil {
			blog.Published.Published, err = time.Parse(time.RFC3339, value./*Clean*/Text())
			if err != nil {
				err = fmt.Errorf("invalid date format, use 2006-01-02 or RFC3339")
			}
		}
	case "revised":
		timeRef := func(t time.Time, err error) (*time.Time, error) {
			return &t, err
		}
		blog.Published.Revised, err = timeRef(time.Parse("2006-01-02", value./*Clean*/Text()))
		if err != nil {
			blog.Published.Revised, err = timeRef(time.Parse(time.RFC3339, value./*Clean*/Text()))
			if err != nil {
				err = fmt.Errorf("invalid date format, use 2006-01-02 or RFC3339")
			}
		}
	case "est-reading":
		blog.EstReading, err = strconv.Atoi(value./*Clean*/Text())
	case "series":
		// @todo
	case "enable-revision-warning":
		switch value.Text() {
		default:
			err = fmt.Errorf("invalid option `%s` for enable-revision-warning, expected one of true, false", value.Text())
		case "true":
			blog.EnableRevisionWarning = true
		case "false":
			blog.EnableRevisionWarning = false
		}
	}
	return
}

type (
	Stack[T any] []T
	Maybe[T any] struct {
		HasValue bool
		Value T
	}
)

func (s Stack[T]) Push(v T) Stack[T] {
	return append(s, v)
}

func (s Stack[T]) Pop() (Stack[T], T) {
	l := len(s)
	Assert(l > 0, "Pop called on empty stack (maybe you want to use SafePop?)")
	return s[:l-1], s[l-1]
}

func (s Stack[T]) SafePop() (Stack[T], Maybe[T]) {
	l := len(s)
	if l > 0 {
		return s[:l-1], Maybe[T]{true, s[l-1]}
	}
	return s, Maybe[T]{HasValue: false}
}

func (s Stack[T]) Peek() T {
	l := len(s)
	return s[l-1]
}

func (s Stack[T]) Empty() (empty Stack[T]) {
	return
}
