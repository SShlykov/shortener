package pg

type Key string

// i would use pgxscan for scanning rows
// like this:
// func (p *Postgres) ScanAllContext(...) error {
//    doLog(...)
//    row, err := pgxscan.QueryContext(...)
//
//
// 	  return pgxscan.ScanAll(...)
// }

const (
	TxKey Key = "tx"
)
