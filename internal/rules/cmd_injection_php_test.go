package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// A SQL string is not a shell command, however many lines it spans.
func TestCmdInjectionPHPMultilineStrings(t *testing.T) {
	runRuleCases(t, CmdInjection, finding.CmdInjection, "CWE-78", []ruleCase{
		{"multi-line double-quoted sql", mkFile("plugins/Actions/ActionReports.php",
			`+$sql = "SELECT MAX(r.idaction)`,
			"+         FROM `$table` r",
			`+         WHERE r.idsite = ? AND r.name IN (%1\$s)";`,
			`+$rows = Db::fetchAll($sql, [$idsite]);`,
		), nil, ""},
		{"multi-line single-quoted sql", mkFile("core/Updates/6.1.0-b1.php",
			`+$sql = 'ALTER TABLE x`,
			"+         ADD COLUMN `$col` INT';",
		), nil, ""},
		{"heredoc sql", mkFile("src/Repo.php",
			`+$sql = <<<SQL`,
			"+SELECT x FROM `$table` WHERE a = ?",
			`+SQL;`,
		), nil, ""},
		{"block comment", mkFile("src/Repo.php",
			`+/*`,
			"+ run `ls $dir` here",
			`+ */`,
		), nil, ""},
		{"backticks after a closed multi-line string", mkFile("public/ls.php",
			`+$sql = "SELECT a`,
			`+         FROM t";`,
			"+$out = `ls $dir`;",
		), []int{3}, finding.Medium},
		{"exec sink on the line after a multi-line string", mkFile("public/ls.php",
			`+$sql = "SELECT a`,
			`+         FROM t";`,
			`+system("ls " . $_GET['d']);`,
		), []int{3}, finding.High},
	})
}
