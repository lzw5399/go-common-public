package transaction

type DummyTransaction struct {
}

func NewDummyTransaction() *DummyTransaction {
	return &DummyTransaction{}
}

func (d *DummyTransaction) GetTransaction() interface{} {
	return nil
}

func (d *DummyTransaction) Commit() error {
	return nil
}

func (d *DummyTransaction) Rollback() error {
	return nil
}
