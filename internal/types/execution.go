package types

type Execution struct {
	ID        string
	TriggerId int
	Status    string
	Trigger   Trigger
}
