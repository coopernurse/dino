package dino

type Provider interface {
	List(domainName string) ([]Record, error)
	Put(rec Record) error
	Delete(domainName string, id string) error
}

type RecordType string

const (
	RecordTypeA     RecordType = "A"
	RecordTypeAAAA             = "AAAA"
	RecordTypeANAME            = "ANAME"
	RecordTypeCNAME            = "CNAME"
	RecordTypeMX               = "MX"
	RecordTypeNS               = "NS"
	RecordTypeSRV              = "SRV"
	RecordTypeTXT              = "TXT"
)

type Record struct {
	Id       string
	Domain   string
	Host     string
	Type     RecordType
	Answer   string
	Ttl      uint32
	Priority uint32
}
