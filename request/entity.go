package request

type Result struct {
	Data            interface{}
	Page            int
	RecordsFiltered interface{}
	RecordsTotal    int
	ResultCode      int
	ResultDesc      string
}
