package dcrm

type NotaryService struct {
	NotaryShareArg *NotaryShareArg
	Notaries       map[string]*NotatoryInfo
}
