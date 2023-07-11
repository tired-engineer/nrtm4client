package persist

type Repository interface {
	InitializeConnectionPool(dbUrl string)
	GetState(string) (NRTMState, *ErrNrtmClient)
	SaveState(NRTMState) error
}
