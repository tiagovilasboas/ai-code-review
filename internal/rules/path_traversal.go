package rules

import (
	"regexp"
	"strings"

	"github.com/tiagovilasboas/ai-code-review/internal/diff"
	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// pathArgs selects which call arguments carry the file path.
type pathArgs int

const (
	firstArg pathArgs = iota
	allArgs
	thirdArg
)

// fileSink is a call that reads, writes, serves or deletes a file by path.
type fileSink struct {
	lang lang
	re   *regexp.Regexp
	args pathArgs
}

var fileSinks = []fileSink{
	{langJS, regexp.MustCompile(`\b(?:fs|fsp|fsPromises|fs\.promises)\.(?:readFile|readFileSync|writeFile|writeFileSync|` +
		`appendFile|appendFileSync|createReadStream|createWriteStream|unlink|unlinkSync|rm|rmSync|` +
		`readdir|readdirSync|open|openSync)\s*\(`), firstArg},
	{langJS, regexp.MustCompile(`\bres\.(?:sendFile|download)\s*\(`), firstArg},
	{langPython, regexp.MustCompile(`(?:^|[^.\w])open\s*\(|\b(?:io|os)\.open\s*\(|\bos\.(?:remove|unlink)\s*\(|` +
		`\bsend_file\s*\(`), firstArg},
	{langPython, regexp.MustCompile(`\bshutil\.(?:copy|copyfile|move|rmtree)\s*\(`), allArgs},
	{langGo, regexp.MustCompile(`\b(?:os\.(?:Open|OpenFile|ReadFile|WriteFile|Create|Remove|RemoveAll)|` +
		`ioutil\.(?:ReadFile|WriteFile)|c\.File)\s*\(`), firstArg},
	{langGo, regexp.MustCompile(`\bhttp\.ServeFile\s*\(`), thirdArg},
	{langJava, regexp.MustCompile(`\bnew\s+(?:File|FileInputStream|FileOutputStream|FileReader|FileWriter|RandomAccessFile)\s*\(|` +
		`\b(?:Paths\.get|Path\.of)\s*\(`), allArgs},
	{langJava, regexp.MustCompile(`\bFiles\.(?:readAllBytes|readString|readAllLines|lines|newInputStream|newOutputStream|` +
		`newBufferedReader|write|writeString|delete|deleteIfExists)\s*\(`), firstArg},
	{langPHP, regexp.MustCompile(`(?:^|[^\w$>:])(?:file_get_contents|file_put_contents|fopen|readfile|unlink|file|` +
		`include|include_once|require|require_once)\s*\(`), firstArg},
}

var (
	phpIncludeBare = regexp.MustCompile(`^\s*(?:include|require)(?:_once)?\s+([^;(][^;]*)`)
	// A path guard needs both a normalization and a base-directory prefix check,
	// or a basename that drops every directory part.
	pathNormalize = regexp.MustCompile(`\bpath\.(?:resolve|normalize)\s*\(|\bos\.path\.(?:realpath|abspath|normpath)\s*\(|` +
		`\.resolve\s*\(|\bfilepath\.(?:Clean|Abs|EvalSymlinks)\s*\(|\brealpath\s*\(|\.normalize\s*\(\)|` +
		`\.getCanonicalPath\s*\(|\.toRealPath\s*\(`)
	pathPrefix = regexp.MustCompile(`\.startsWith\s*\(|\.startswith\s*\(|\bstrings\.HasPrefix\s*\(|\bfilepath\.Rel\s*\(|` +
		`\.is_relative_to\s*\(|\bstr_starts_with\s*\(|\bstrpos\s*\([^)]*\)\s*[!=]==?\s*0|\bos\.path\.commonpath\s*\(`)
	pathBasename = regexp.MustCompile(`\bpath\.basename\s*\(|\bos\.path\.basename\s*\(|\bfilepath\.Base\s*\(|` +
		`(?:^|[^.\w])basename\s*\(|\bFilenameUtils\.getName\s*\(|\bsecure_filename\s*\(`)
	sendFileRoot = regexp.MustCompile(`\broot\s*:`)
)

// PathTraversal flags file reads, writes and sends whose path visibly comes
// from request input with no normalization plus base-directory check (or
// basename) in the same hunk (CWE-22). Paths not from request input never fire.
func PathTraversal(f diff.File) []finding.Finding {
	lg := langOf(f.Path)
	if lg == langOther {
		return nil
	}
	php := lg == langPHP
	tainted := taintedNames(f.Lines, php)
	safe := allowlistedValues(f.Lines)
	var out []finding.Finding
	for i, l := range f.Lines {
		if l.Kind != diff.Added || IsCommentLine(l.Text) {
			continue
		}
		path, call, ok := filePathArg(f.Lines, i, lg)
		if !ok || !reachesRequest(maskAllowlisted(path, safe), tainted, php) {
			continue
		}
		if sendFileRoot.MatchString(call) || pathGuarded(guardScope(f.Lines, i, lg)) {
			continue
		}
		out = add(out, finding.Input{
			Title:    "File path from request input",
			Class:    finding.PathTraversal,
			Path:     f.Path,
			Line:     l.Line,
			CWE:      "CWE-22",
			OWASP:    finding.LabelsFor(finding.PathTraversal),
			Severity: finding.High,
		})
	}
	return out
}

// filePathArg returns the path argument(s) and the call text of the first
// file sink on lines[index].
func filePathArg(lines []diff.Line, index int, lg lang) (path, call string, ok bool) {
	text := lines[index].Text
	if lg == langPHP {
		if m := phpIncludeBare.FindStringSubmatch(text); m != nil {
			return m[1], m[1], true
		}
	}
	for _, s := range fileSinks {
		loc := s.re.FindStringIndex(text)
		if s.lang != lg || loc == nil {
			continue
		}
		call = callFrom(lines, index, loc[0])
		args := callArgs(call, loc[1]-1-loc[0])
		switch {
		case len(args) == 0:
			continue
		case s.args == allArgs:
			return strings.Join(args, ","), call, true
		case s.args == thirdArg && len(args) >= 3:
			return args[2], call, true
		case s.args == firstArg:
			return args[0], call, true
		}
	}
	return "", "", false
}

// guardScope is where a guard must sit: the enclosing function for brace
// languages, the whole contiguous hunk for Python. Both stop at the hunk.
func guardScope(lines []diff.Line, index int, lg lang) []diff.Line {
	if lg == langPython {
		return hunkAround(lines, index)
	}
	return enclosingHandler(lines, index)
}

// pathGuarded reports a visible guard in the hunk: normalization plus a
// base-directory prefix check, or a basename.
func pathGuarded(hunk []diff.Line) bool {
	normalized, prefixed := false, false
	for _, l := range hunk {
		if IsCommentLine(l.Text) {
			continue
		}
		if pathBasename.MatchString(l.Text) {
			return true
		}
		normalized = normalized || pathNormalize.MatchString(l.Text)
		prefixed = prefixed || pathPrefix.MatchString(l.Text)
	}
	return normalized && prefixed
}
