package rules

import (
	"testing"

	"github.com/tiagovilasboas/ai-code-review/internal/finding"
)

// Constant pieces, table-prefix helpers and placeholder lists carry no request data.
func TestSQLInjectionConstantConcatenation(t *testing.T) {
	runRuleCases(t, SQLInjection, finding.SQLInjection, "CWE-89", []ruleCase{
		{"prefixTable with class constant and binds", mkFile("plugins/Rec.php",
			`+Db::query('INSERT INTO ' . Common::prefixTable(self::TABLE) . ' (a,b) VALUES (?,?)', [$x, $y]);`,
		), nil, ""},
		{"static constant as table", mkFile("plugins/Rec.php",
			`+Db::query('SELECT a FROM ' . static::TABLE . ' WHERE id = ?', [$x]);`,
		), nil, ""},
		{"UPPER_CASE constant as table", mkFile("plugins/Rec.php",
			`+Db::query('SELECT a FROM ' . TABLE_NAME . ' WHERE id = ?', [$x]);`,
		), nil, ""},
		{"getSqlStringFieldsArray placeholder list", mkFile("plugins/Rec.php",
			`+Db::query('SELECT * FROM ' . Common::prefixTable('log_visit') . ' WHERE id IN (' . Common::getSqlStringFieldsArray($ids) . ')', $ids);`,
		), nil, ""},
		{"placeholder helper method", mkFile("plugins/Rec.php",
			`+$db->query('SELECT a FROM t WHERE id IN (' . $this->placeholders($ids) . ')', $ids);`,
		), nil, ""},
		{"built sql var from prefix helper", mkFile("plugins/Rec.php",
			`+$sql = 'SELECT a FROM ' . Common::prefixTable('x') . ' WHERE idsite = ?';`,
			`+$rows = Db::query($sql, [$idsite]);`,
		), nil, ""},
		{"placeholder var from implode array_fill", mkFile("plugins/Rec.php",
			`+$ph = implode(',', array_fill(0, count($ids), '?'));`,
			`+Db::query("SELECT a FROM t WHERE id IN ($ph)", $ids);`,
		), nil, ""},
		{"placeholder var from go strings.Repeat", mkFile("internal/store/orders.go",
			`+	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")`,
			`+	rows, err := db.Query("SELECT id FROM orders WHERE id IN ("+ph+")", args...)`,
		), nil, ""},
		{"request variable next to a prefix helper", mkFile("plugins/Rec.php",
			`+Db::query('SELECT a FROM ' . Common::prefixTable('t') . ' WHERE id = ' . $_GET['id']);`,
		), []int{1}, finding.High},
		{"request variable next to a placeholder list", mkFile("plugins/Rec.php",
			`+Db::query('SELECT a FROM t WHERE id IN (' . Common::getSqlStringFieldsArray($ids) . ') AND n = ' . $name, $ids);`,
		), []int{1}, finding.Medium},
		{"prefixTable of a variable is not proven", mkFile("plugins/Rec.php",
			`+Db::query('SELECT a FROM ' . Common::prefixTable($_GET['t']) . ' WHERE id = ?', [$x]);`,
		), []int{1}, finding.High},
		{"static property is not a constant", mkFile("plugins/Rec.php",
			`+Db::query('SELECT a FROM ' . self::$table . ' WHERE id = ?', [$x]);`,
		), []int{1}, finding.Medium},
	})
}
