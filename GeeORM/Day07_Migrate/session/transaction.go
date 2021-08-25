package session

import "geeorm/log"

/*
	封装事务相关，统一打印日志
 */
func (s *Session) Begin() (err error) {
	log.Info("[Begin] transaction begin")
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error("[Begin] transaction begin err:", err)
		return
	}
	return
}

func (s *Session) Commit() (err error) {
	log.Info("[Commit] transaction commit")
	if err = s.tx.Commit(); err != nil {
		log.Error("[Commit] transaction commit err:", err)
		return
	}
	return
}

func (s *Session) Rollback() (err error) {
	log.Info("[Rollback] transaction rollback")
	if err := s.tx.Rollback(); err != nil {
		log.Error("[Rollback] transaction rollback err:", err)
	}
	return
}
