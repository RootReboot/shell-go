package token

type TokenType int

const (
	//EOF - End of file
	TokenEOF TokenType = iota
	TokenWord
	TokenPipe
	TokenRedirectOut
	TokenRedirectErr
	TokenAppendRedirectOut
	TokenAppendRedirectErr
)
