package holder

import (
	"fmt"
)

import (
	"github.com/dk-lockdown/seata-golang/tc/model"
	"github.com/dk-lockdown/seata-golang/tc/session"
)

type LogOperation byte

const (
	LogOperationGlobalAdd LogOperation = iota
	/**
	 * Global update log operation.
	 */
	LogOperationGlobalUpdate
	/**
	 * Global remove log operation.
	 */
	LogOperationGlobalRemove
	/**
	 * Branch add log operation.
	 */
	LogOperationBranchAdd
	/**
	 * Branch update log operation.
	 */
	LogOperationBranchUpdate
	/**
	 * Branch remove log operation.
	 */
	LogOperationBranchRemove
)

func (t LogOperation) String() string {
	switch t {
	case LogOperationGlobalAdd:
		return "GlobalAdd"
	case LogOperationGlobalUpdate:
		return "GlobalUpdate"
	case LogOperationGlobalRemove:
		return "GlobalRemove"
	case LogOperationBranchAdd:
		return "BranchAdd"
	case LogOperationBranchUpdate:
		return "BranchUpdate"
	case LogOperationBranchRemove:
		return "BranchRemove"
	default:
		return fmt.Sprintf("%d", t)
	}
}

type ITransactionStoreManager interface {
	/**
	 * Write session boolean.
	 *
	 * @param logOperation the log operation
	 * @param session      the session
	 * @return the boolean
	 */
	WriteSession(logOperation LogOperation, session session.SessionStorable) bool


	/**
	 * Read global session global session.
	 *
	 * @param xid the xid
	 * @return the global session
	 */
	ReadSession(xid string) *session.GlobalSession

	/**
	 * Read session global session.
	 *
	 * @param xid the xid
	 * @param withBranchSessions the withBranchSessions
	 * @return the global session
	 */
	ReadSessionWithBranchSessions(xid string, withBranchSessions bool) *session.GlobalSession

	/**
	 * Read session by status list.
	 *
	 * @param sessionCondition the session condition
	 * @return the list
	 */
	ReadSessionWithSessionCondition(sessionCondition model.SessionCondition) []*session.GlobalSession

	/**
	 * Shutdown.
	 */
	Shutdown()

	/**
	 * Gets current max session id.
	 *
	 * @return the current max session id
	 */
	GetCurrentMaxSessionId() int64
}

type AbstractTransactionStoreManager struct {

}

func (transactionStoreManager *AbstractTransactionStoreManager) WriteSession(logOperation LogOperation, session session.SessionStorable) bool {
	return true
}



func (transactionStoreManager *AbstractTransactionStoreManager) ReadSession(xid string) *session.GlobalSession {
	return nil
}


func (transactionStoreManager *AbstractTransactionStoreManager) ReadSessionWithBranchSessions(xid string, withBranchSessions bool) *session.GlobalSession {
	return nil
}

func (transactionStoreManager *AbstractTransactionStoreManager) ReadSessionWithSessionCondition(sessionCondition model.SessionCondition) []*session.GlobalSession {
	return nil
}

func (transactionStoreManager *AbstractTransactionStoreManager) Shutdown() {

}


func (transactionStoreManager *AbstractTransactionStoreManager) GetCurrentMaxSessionId() int64 {
	return 0
}


