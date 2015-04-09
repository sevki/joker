package analyzers

// Github comment structure.
type Message struct {
	Body     string
	Filename string
	Line     int
	// Github doesn't care about this in commits.
	Col int
}

// Scanner interface for the analyzerss
type Scanner interface {
	Scan() bool
	Message() Message
}

var analysers map[string]InitFunc

func init() {
	analysers = make(map[string]InitFunc)
}

type InitFunc func() Scanner

func GetScanner(scnr string) Scanner {
	a := analysers[scnr]()
	return a
}
func Register(scnr string, scnrFunc InitFunc) {
	analysers[scnr] = scnrFunc
}
