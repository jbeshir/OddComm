package irc

// Maps characters to strings, and visa versa.
// Useful for implementing extbans and opstrs.
type CMapper struct {
	charToStr map[int]string
	strToChar map[string]int
}

// NewCMapper returns a new CMapper, ready to use.
func NewCMapper() (m *CMapper) {
	m = new(CMapper)
	m.charToStr = make(map[int]string)
	m.strToChar = make(map[string]int)
	return
}

// Add adds an str to this str mapper.
// Must only be used during init().
func (m *CMapper) Add(char int, str string) {
	m.charToStr[char] = str
	m.strToChar[str] = char
}

// Str looks up the string for the given character.
// It returns "" for none.
func (m *CMapper) Str(char int) string {
	return m.charToStr[char]
}

// Char looks up the character for the given string.
// It returns 0 for none.
func (m *CMapper) Char(str string) int {
	return m.strToChar[str]
}
