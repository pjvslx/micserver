package roc

import (
	"sync"
)

const (
	CACHE_POOL_GROUP_SUM = 8
)

type catchServerInfo struct {
	moduleid string
}

type serverInfoMap map[string]*catchServerInfo
type objIDToServerMap map[string]*catchServerInfo

type Cache struct {
	catchServer serverInfoMap
	catchType   map[ROCObjType]objIDToServerMap
	mutex       sync.Mutex
}

var _gCache Cache

func GetCache() *Cache {
	return &_gCache
}

func (this *Cache) catchGetTypeMust(objType ROCObjType) objIDToServerMap {
	if this.catchType == nil {
		this.catchType = make(map[ROCObjType]objIDToServerMap)
	}
	if v, ok := this.catchType[objType]; !ok {
		v = make(objIDToServerMap)
		this.catchType[objType] = v
		return v
	} else {
		return v
	}
}

func (this *Cache) catchGetServerMust(moduleid string) *catchServerInfo {
	if this.catchServer == nil {
		this.catchServer = make(serverInfoMap)
	}
	if v, ok := this.catchServer[moduleid]; !ok {
		v = &catchServerInfo{
			moduleid: moduleid,
		}
		this.catchServer[moduleid] = v
		return v
	} else {
		return v
	}
}

func (this *Cache) Set(objType ROCObjType, objID string, moduleid string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	m := this.catchGetTypeMust(objType)
	server := this.catchGetServerMust(moduleid)
	m[objID] = server
}

func (this *Cache) SetM(objType ROCObjType, objIDs []string, moduleid string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	m := this.catchGetTypeMust(objType)
	server := this.catchGetServerMust(moduleid)
	for _, v := range objIDs {
		m[v] = server
	}
}

func (this *Cache) Del(objType ROCObjType, objID string, moduleid string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	m := this.catchGetTypeMust(objType)
	if info, ok := m[objID]; ok && info.moduleid == moduleid {
		delete(m, objID)
	}
}

func (this *Cache) DelM(objType ROCObjType, objIDs []string, moduleid string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	m := this.catchGetTypeMust(objType)
	for _, v := range objIDs {
		if info, ok := m[v]; ok && info.moduleid == moduleid {
			delete(m, v)
		}
	}
}

func (this *Cache) Get(objType ROCObjType, objID string) string {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	m := this.catchGetTypeMust(objType)
	if v, ok := m[objID]; ok && v != nil {
		return v.moduleid
	}
	return ""
}
