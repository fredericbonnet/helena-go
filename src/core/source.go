//
// Helena sources
//

package core

//
// Source descriptor
//
type Source struct {
	// Path of file-backed sources
	Filename *string

	// Source content (optional for files)
	Content *string
}

//
// Position in source
//
type SourcePosition struct {
	// Character Index (zero-indexed)
	Index uint

	// Line number (zero-indexed)
	Line uint

	// Column number (zero-indexed)
	Column uint
}
