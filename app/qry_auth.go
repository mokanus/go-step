package app

import "sync"

var (
	qryAuthUserMap          = make(map[string]string)
	qryAuthSessionMap       = make(map[string]*qryAuthSession)
	qryAuthSessionMapLocker = new(sync.RWMutex)
)

type qryAuthSession struct {
	Token    string
	Username string
	Expired  uint32
}

func addQryAuthSession(session *qryAuthSession) {
	qryAuthSessionMapLocker.Lock()
	defer qryAuthSessionMapLocker.Unlock()
	qryAuthSessionMap[session.Token] = session
}

func getQryAuthSession(token string) *qryAuthSession {
	qryAuthSessionMapLocker.RLock()
	defer qryAuthSessionMapLocker.RUnlock()
	return qryAuthSessionMap[token]
}

func CheckQryAuthUser(username string, password string) bool {
	expectedPassword, ok := qryAuthUserMap[username]
	if !ok || password != expectedPassword {
		return false
	}
	return true
}
