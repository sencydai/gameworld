package engine

import (
	"database/sql"
	"sync"
	"time"

	//
	_ "github.com/go-sql-driver/mysql"

	"github.com/json-iterator/go"
	g "github.com/sencydai/gameworld/gconfig"
	t "github.com/sencydai/gameworld/typedefine"

	"github.com/sencydai/gameworld/log"
)

type actorBuffer struct {
	ActorName  string
	AccountId  int
	Camp       int
	Sex        int
	Level      int
	Power      int
	LoginTime  time.Time
	LogoutTime time.Time
	BaseData   []byte
	ExData     []byte
}

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	engine *sql.DB

	stmtSysMutex sync.Mutex

	stmtAccount      *sql.Stmt
	stmtAccountMutex sync.Mutex

	stmtAccountActors      *sql.Stmt
	stmtAccountActorsMutex sync.Mutex

	stmtActorCount *sql.Stmt

	stmtInsertActor *sql.Stmt

	stmtQueryActor      *sql.Stmt
	stmtQueryCacheActor *sql.Stmt

	actorBuffers  = make(map[int64]*actorBuffer)
	actorBufferMu sync.RWMutex
)

func InitEngine() {
	var err error
	if engine, err = sql.Open("mysql", g.GameConfig.Database); err != nil {
		panic(err)
	} else if err = engine.Ping(); err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case <-time.After(time.Hour):
				if err := engine.Ping(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	stmtAccount, err = engine.Prepare("select accountid,password,gmlevel from account where accountname=?")
	if err != nil {
		panic(err)
	}

	stmtAccountActors, err = engine.Prepare("select actorid,actorname,camp,sex,level from actor where accountid=? order by level desc")
	if err != nil {
		panic(err)
	}

	stmtActorCount, err = engine.Prepare("select count(actorid) from actor where accountid=?")
	if err != nil {
		panic(err)
	}

	stmtInsertActor, err = engine.Prepare("insert actor values(?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		panic(err)
	}

	stmtQueryActor, err = engine.Prepare("select actorname,accountid,camp,sex,level,power,logintime,logouttime,basedata,exdata from actor where actorid=?")
	if err != nil {
		panic(err)
	}

	stmtQueryCacheActor, err = engine.Prepare("select actorname,accountid,camp,sex,level,power,logintime,logouttime,basedata from actor where actorid=?")
	if err != nil {
		panic(err)
	}
}

func GetMaxActorId() (int64, error) {
	var maxId sql.NullInt64
	err := engine.QueryRow("select max(actorid) from actor where serverid=?", g.GameConfig.ServerId).Scan(&maxId)
	if err != nil {
		return 0, err
	}
	return maxId.Int64, nil
}

func GetAllActorNames() (map[string]int64, error) {
	rows, err := engine.Query("select actorid,actorname from actor")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		actorid   int64
		actorname string
	)
	names := make(map[string]int64)
	for rows.Next() {
		if err := rows.Scan(&actorid, &actorname); err != nil {
			return nil, err
		}
		names[actorname] = actorid
	}

	return names, nil
}

func GetSystemData(index int, defValue string) (string, error) {
	var id int
	var data sql.NullString
	err := engine.QueryRow("select id,data from sysdata where id=?", index).Scan(&id, &data)
	if err != nil {
		if err != sql.ErrNoRows {
			panic(err)
		}
		_, err = engine.Exec("insert into sysdata(id,data) values(?,?)", index, defValue)
		return defValue, err
	}

	if !data.Valid {
		return defValue, err
	}
	return data.String, nil
}

func FlushSystemData(values map[int]string) {
	stmtSysMutex.Lock()
	defer stmtSysMutex.Unlock()

	tx, err := engine.Begin()
	if err != nil {
		log.Errorf("FlushSystemData begin error: %s", err.Error())
		return
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare("update sysdata set data = ? where id = ?")
	if err != nil {
		log.Errorf("FlushSystemData prepare error: %s", err.Error())
		return
	}

	for index, text := range values {
		_, err := stmt.Exec(text, index)
		if err != nil {
			log.Errorf("FlushSystemData index(%d) error: %s", index, err.Error())
		}
	}

	stmt.Close()

	err = tx.Commit()
	if err != nil {
		log.Errorf("FlushSystemData commit error: %s", err.Error())
	}
}

func GetAccountInfo(name string) (int, string, byte, error) {
	stmtAccountMutex.Lock()
	defer stmtAccountMutex.Unlock()

	var (
		accountid int
		password  string
		gmlevel   byte
	)
	err := stmtAccount.QueryRow(name).Scan(&accountid, &password, &gmlevel)
	return accountid, password, gmlevel, err
}

func GetAccountActors(accountId int) ([]*t.AccountActor, error) {
	stmtAccountActorsMutex.Lock()
	defer stmtAccountActorsMutex.Unlock()

	rows, err := stmtAccountActors.Query(accountId)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}

	defer rows.Close()

	actors := make([]*t.AccountActor, 0)
	for rows.Next() {
		actor := &t.AccountActor{}
		if err := rows.Scan(&actor.ActorId, &actor.ActorName, &actor.Camp, &actor.Sex, &actor.Level); err != nil {
			return nil, err
		}
		buff := getActorBuffer(int64(actor.ActorId))
		if buff != nil {
			actor.ActorName = buff.ActorName
			actor.Level = buff.Level
		}
		actors = append(actors, actor)
	}

	return actors, nil
}

func GetActorCount(accountId int) (int, error) {
	var result sql.NullInt64
	err := stmtActorCount.QueryRow(accountId).Scan(&result)
	if err != nil {
		return 0, err
	}
	return int(result.Int64), nil
}

func InsertActor(actor *t.Actor) error {
	baseData, err := json.MarshalIndent(actor.BaseData, "", " ")
	if err != nil {
		return err
	}
	exData, err := json.MarshalIndent(actor.ExData, "", " ")
	if err != nil {
		return err
	}
	_, err = stmtInsertActor.Exec(actor.ActorId, actor.ActorName, actor.AccountId,
		g.GameConfig.ServerId, actor.Camp, actor.Sex, actor.Level, actor.Power,
		actor.CreateTime, actor.LoginTime, actor.LogoutTime, string(baseData), string(exData))
	return err
}

func QueryActor(actorId int64) (*t.Actor, error) {
	actor := &t.Actor{ActorId: actorId}
	buff := getActorBuffer(actorId)
	if buff != nil {
		actor.ActorName = buff.ActorName
		actor.AccountId = buff.AccountId
		actor.Camp = buff.Camp
		actor.Sex = buff.Sex
		actor.Level = buff.Level
		actor.Power = buff.Power
		actor.LoginTime = buff.LoginTime
		actor.LogoutTime = buff.LogoutTime
		actor.BaseData = &t.ActorBaseData{}
		json.Unmarshal(buff.BaseData, actor.BaseData)
		actor.ExData = &t.ActorExData{}
		json.Unmarshal(buff.ExData, actor.ExData)
		return actor, nil
	}

	var (
		baseData sql.NullString
		exData   sql.NullString
	)
	err := stmtQueryActor.QueryRow(actorId).Scan(&actor.ActorName, &actor.AccountId, &actor.Camp,
		&actor.Sex, &actor.Level, &actor.Power, &actor.LoginTime, &actor.LogoutTime, &baseData, &exData)
	if err != nil {
		return nil, err
	}
	actor.BaseData = &t.ActorBaseData{}
	actor.ExData = &t.ActorExData{}
	if baseData.Valid {
		if err = json.Unmarshal([]byte(baseData.String), actor.BaseData); err != nil {
			return nil, err
		}
	}
	if exData.Valid {
		if err = json.Unmarshal([]byte(exData.String), actor.ExData); err != nil {
			return nil, err
		}
	}
	return actor, nil
}

func QueryActorCache(actorId int64) (*t.Actor, error) {
	actor := &t.Actor{ActorId: actorId}
	buff := getActorBuffer(actorId)
	if buff != nil {
		actor.ActorName = buff.ActorName
		actor.AccountId = buff.AccountId
		actor.Camp = buff.Camp
		actor.Sex = buff.Sex
		actor.Level = buff.Level
		actor.Power = buff.Power
		actor.LoginTime = buff.LoginTime
		actor.LogoutTime = buff.LogoutTime
		actor.BaseData = &t.ActorBaseData{}
		json.Unmarshal(buff.BaseData, actor.BaseData)
		return actor, nil
	}

	var (
		baseData sql.NullString
	)
	err := stmtQueryCacheActor.QueryRow(actorId).Scan(&actor.ActorName, &actor.AccountId, &actor.Camp,
		&actor.Sex, &actor.Level, &actor.Power, &actor.LoginTime, &actor.LogoutTime, &baseData)
	if err != nil {
		return nil, err
	}
	actor.BaseData = &t.ActorBaseData{}
	if baseData.Valid {
		if err = json.Unmarshal([]byte(baseData.String), actor.BaseData); err != nil {
			return nil, err
		}
	}
	return actor, nil
}

func UpdateActor(actor *t.Actor) {
	actorBufferMu.Lock()
	defer actorBufferMu.Unlock()

	baseData, err := json.MarshalIndent(actor.BaseData, "", " ")
	if err != nil {
		log.Errorf("marshal actor(%d) baseData error: %s", actor.ActorId, err.Error())
		return
	}

	exData, err := json.MarshalIndent(actor.ExData, "", " ")
	if err != nil {
		log.Errorf("marshal actor(%d) exData error: %s", actor.ActorId, err.Error())
		return
	}
	if buff, ok := actorBuffers[actor.ActorId]; ok {
		buff.ActorName = actor.ActorName
		// buff.AccountId = actor.AccountId
		// buff.Camp = actor.Camp
		// buff.Sex = actor.Sex
		buff.Level = actor.Level
		buff.Power = actor.Power
		buff.LoginTime = actor.LoginTime
		buff.LogoutTime = actor.LogoutTime
		buff.BaseData = baseData
		buff.ExData = exData
	} else {
		actorBuffers[actor.ActorId] = &actorBuffer{
			ActorName:  actor.ActorName,
			AccountId:  actor.AccountId,
			Camp:       actor.Camp,
			Sex:        actor.Sex,
			Level:      actor.Level,
			Power:      actor.Power,
			LoginTime:  actor.LoginTime,
			LogoutTime: actor.LogoutTime,
			BaseData:   baseData,
			ExData:     exData,
		}
	}
}

func getActorBuffer(actorId int64) *actorBuffer {
	actorBufferMu.RLock()
	defer actorBufferMu.RUnlock()

	return actorBuffers[actorId]
}

func FlushActorBuffers() {
	actorBufferMu.Lock()
	defer actorBufferMu.Unlock()
	if len(actorBuffers) == 0 {
		return
	}

	tx, err := engine.Begin()
	if err != nil {
		log.Errorf("FlushActorBuffers error: %s", err.Error())
		return
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare("update actor set actorname=?,level=?,power=?,logintime=?,logouttime=?,basedata=?,exdata=? where actorid=?")
	if err != nil {
		log.Errorf("FlushActorBuffers prepare error: %s", err.Error())
		return
	}

	for actorId, buff := range actorBuffers {
		//actorName, level, power, logintime, logouttime, string(baseData), string(exData), actorId
		_, err := stmt.Exec(buff.ActorName, buff.Level, buff.Power, buff.LoginTime, buff.LogoutTime, string(buff.BaseData), string(buff.ExData), actorId)
		if err != nil {
			log.Errorf("save actor(%d) data error: %s", actorId, err.Error())
		}
	}

	stmt.Close()

	err = tx.Commit()
	if err != nil {
		log.Errorf("FlushActorBuffers commit error: %s", err.Error())
	} else {
		for actorId := range actorBuffers {
			delete(actorBuffers, actorId)
		}
		//actorBuffers = make(map[int64]*actorBuffer)
	}
}

func GetCacheCount() int {
	actorBufferMu.RLock()
	defer actorBufferMu.RUnlock()

	return len(actorBuffers)
}
