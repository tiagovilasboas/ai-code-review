package rules

import (
	"path"
	"regexp"
	"strings"
)

var (
	// translationPath matches message catalogs: lang/, locales/, i18n/, l10n/,
	// translations/ directories and gettext files.
	translationPath = regexp.MustCompile(`(?i)(?:^|/)(?:lang|locales?|i18n|l10n|translations?)/|\.pot?$`)
	// translationValue is a line that holds one plain string value: a JSON or
	// YAML "key": "text" entry, a PHP 'key' => 'text' entry, or a gettext
	// msgid/msgstr line (and its continuation).
	translationValue = []*regexp.Regexp{
		regexp.MustCompile(`^\s*"[^"]+"\s*:\s*"(?:[^"\\]|\\.)*"\s*,?\s*$`),
		regexp.MustCompile(`^\s*[\w.\-]+\s*:\s+(?:"(?:[^"\\]|\\.)*"|'[^']*'|[^\s"'\[{|>#-][^#]*)\s*$`),
		regexp.MustCompile(`^\s*['"][^'"]+['"]\s*=>\s*(?:"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')\s*,?\s*$`),
	}
	gettextLine = regexp.MustCompile(`^\s*(?:msgid|msgstr|msgctxt|msgid_plural)(?:\[\d+\])?\s+".*"\s*$|^\s*".*"\s*$`)
)

// translationCopy reports an added line of a message catalog that is only a
// string value: UI copy written for users.
func translationCopy(p, text string) bool {
	if !translationPath.MatchString(p) {
		return false
	}
	if ext := strings.ToLower(path.Ext(p)); (ext == ".po" || ext == ".pot") && gettextLine.MatchString(text) {
		return true
	}
	for _, re := range translationValue {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}
