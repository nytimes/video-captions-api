//nolint:gochecknoglobals
package vtt

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// An error object that holds a line number and component
type ValidatorError struct {
	component string
	message   string
	line      int
}

func (e *ValidatorError) Error() string {
	return fmt.Sprintf("[%s] %s [line %d]", e.component, e.message, e.line)
}

// string constants
const stringArrow = "-->"
const stringNote = "NOTE"
const stringStyle = "STYLE"

// regexp pattern "constants"
var patternSignature = regexp.MustCompile(`\x{feff}?WEBVTT`)
var patternSignatureComment = regexp.MustCompile(`\s+.*`)

var patternMetadata = regexp.MustCompile(`^[A-z_-]*: .*$`)

var patternTimestamp = regexp.MustCompile(`(\d{2}:)?\d{2}:\d{2}\.\d{3}`)
var patternCueArrow = regexp.MustCompile(fmt.Sprintf(`\s+%s\s+`, stringArrow))
var patternCueSettings = regexp.MustCompile(`\s+(\S+:\S+)`)

var patternNote = regexp.MustCompile(fmt.Sprintf(`%s(\s*.+)?`, stringNote))

// a token struct for the parser, useful for valdating complex lines
// with specific error messages
type parserToken struct {
	name     string
	pattern  *regexp.Regexp
	example  string
	message  string
	repeat   bool
	optional bool
}

// Validate will ensure the incoming io.Reader is a valid
// vtt file. If the file is not valid vtt an error will
// be returned.
//
// See the VTT Spec for details about file structure:
// https://www.w3.org/TR/webvtt1/
func Validate(reader io.Reader) error {
	// setup a scanner to read the file line by line
	// by default the scanner will handle \r and \n
	// properly to the VTT spec
	scanner := bufio.NewScanner(reader)
	lineNumber := 0

	// the first block must be the header
	// read one block and ensure it's a proper header
	header := readBlock(scanner)
	lineNumber += len(header) + 1

	if header == nil {
		return errors.New("file is empty")
	}

	err := validateHeader(header)

	if err != nil {
		return err
	}

	// scan the file block by block and check for validity
	for block := readBlock(scanner); block != nil; block = readBlock(scanner) {
		err := validateBlock(block)

		if err != nil {
			// if we get a validation error, the line numbers are relative
			// to the block, so we need to convert them to absolute line numbers
			if verr, ok := err.(*ValidatorError); ok {
				verr.line += lineNumber
			}

			return err
		}

		// increment our line count by the size of the block and the empty blank line
		lineNumber += len(block) + 1
	}

	return nil
}

// readBlock will use the scanner to scan through the file until it encounters
// an empty blank line
func readBlock(scanner *bufio.Scanner) []string {
	block := []string{}
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			return block
		}

		block = append(block, line)
	}

	if len(block) > 0 {
		return block
	}

	return nil
}

// validateHeader will check the signature line as well as the metadata
func validateHeader(block []string) error {
	err := validateSignature(block[0])

	if err != nil {
		return &ValidatorError{
			line:      1,
			component: "header",
			message:   err.Error(),
		}
	}

	block = block[1:]

	for i, line := range block {
		validMetadata := patternMetadata.MatchString(line)

		if !validMetadata && i == 0 {
			return &ValidatorError{
				line:      i + 1,
				component: "header",
				message:   "invalid header: an empty new line is required after the header",
			}
		}

		if !validMetadata {
			return &ValidatorError{
				line:      i + 1,
				component: "header",
				message:   fmt.Sprintf("invalid header: malformed metadata: %s", line),
			}
		}
	}

	return nil
}

// validateBlock is basically a switch statement which decides which validator
// function to use for the based off of characteristics in the first line.
func validateBlock(block []string) error {
	firstLine := block[0]

	if strings.Contains(firstLine, stringArrow) {
		return validateCueBlock(block)
	}

	if strings.Contains(firstLine, stringNote) {
		return validateNoteBlock(block)
	}

	if strings.Contains(firstLine, stringStyle) {
		return validateStyleBlock(block)
	}

	return fmt.Errorf("unknown block type: %s", firstLine)
}

func validateSignature(signature string) error {
	tokens := []*parserToken{
		{
			name:    "signature",
			pattern: patternSignature,
			example: "WEBVTT",
		},
		{
			name:     "signature",
			pattern:  patternSignatureComment,
			message:  "whitespace is required before comment",
			optional: true,
		},
	}

	err := validateTokens(tokens, signature)

	if err != nil {
		return err
	}

	return nil
}

func validateCueBlock(block []string) error {
	tokens := []*parserToken{
		{
			name:    "start timestamp",
			pattern: patternTimestamp,
			example: "00:00:00.000",
		},
		{
			name:    "arrow",
			pattern: patternCueArrow,
			example: " --> ",
		},
		{
			name:    "end timestamp",
			pattern: patternTimestamp,
			example: "00:00:00.000",
		},
		{
			name:     "settings",
			pattern:  patternCueSettings,
			example:  " name:value",
			repeat:   true,
			optional: true,
		},
	}

	timingLine := block[0]
	err := validateTokens(tokens, timingLine)

	if err != nil {
		return &ValidatorError{
			component: "cue",
			line:      1,
			message:   err.Error(),
		}
	}

	return nil
}

func validateNoteBlock(block []string) error {
	match := patternNote.MatchString(block[0])

	if !match {
		return &ValidatorError{
			line:      0,
			component: "note",
			message:   "invalid note",
		}
	}

	return nil
}

func validateStyleBlock(block []string) error {
	cssStr := strings.Join(block[1:], "\n")

	p := css.NewParser(bytes.NewBufferString(cssStr), false)

	for {
		grammar, _, _ := p.Next()

		if grammar == css.ErrorGrammar {
			if p.Err() != io.EOF {
				e := p.Err().(*parse.Error)
				return &ValidatorError{
					component: "style",
					message:   e.Message,
					line:      e.Line + 1,
				}
			}

			return nil
		}
	}
}

// validateTokens creates an easy way to check if a series of
// regexps are valid matches and if not to return a helpful and
// specific error message.
func validateTokens(tokens []*parserToken, str string) error {
	i := 0
	var t *parserToken
	for str != "" && i < len(tokens) {
		t = tokens[i]

		matches := t.pattern.FindStringIndex(str)

		if matches == nil || matches[0] != 0 {
			message := t.message
			if message == "" {
				got := strings.Split(strings.TrimSpace(str), " ")[0]
				message = fmt.Sprintf(`expecting: "%s", got: "%s"`, t.example, got)
			}
			return fmt.Errorf("invalid %s, %s", t.name, message)
		}

		end := matches[1]

		str = str[end:]

		if !t.repeat {
			i++
		}
	}

	// if we didn't get to our last token and that token is not optional
	if str == "" && i+1 != len(tokens) && !t.optional {
		t = tokens[i]
		message := t.message
		if message == "" {
			message = fmt.Sprintf(`expecting: "%s", got: ""`, t.example)
		}
		return fmt.Errorf("invalid %s, %s", t.name, message)
	}

	return nil
}
