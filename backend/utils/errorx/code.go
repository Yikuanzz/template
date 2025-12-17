package errorx

import "sync"

var (
	// codeRegistry 错误码注册表
	codeRegistry = make(map[int32]string)
	// registryMu 保护注册表的互斥锁
	registryMu sync.RWMutex
)

// Register 注册错误码和对应的消息模板
// code: 错误码
// message: 错误消息模板，支持 {key} 占位符
func Register(code int32, message string) {
	registryMu.Lock()
	defer registryMu.Unlock()
	codeRegistry[code] = message
}

// RegisterBatch 批量注册错误码
func RegisterBatch(codes map[int32]string) {
	registryMu.Lock()
	defer registryMu.Unlock()
	for code, message := range codes {
		codeRegistry[code] = message
	}
}

// getRegisteredMessage 获取注册的错误消息
func getRegisteredMessage(code int32) string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return codeRegistry[code]
}

// IsRegistered 检查错误码是否已注册
func IsRegistered(code int32) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, ok := codeRegistry[code]
	return ok
}
