package roc

type IObj interface {
	GetROCObjType() ROCObjType
	GetROCObjID() string
	ROCCall(*ROCPath, []byte) ([]byte, error)
}

type IROCObjEventHook interface {
	OnROCObjAdd(IObj)
	OnROCObjDel(IObj)
}
