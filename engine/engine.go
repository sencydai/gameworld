package engine

import (
	"database/sql"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/json-iterator/go"
	"github.com/sencydai/gameworld/gconfig"
	. "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils"
	"github.com/sencydai/utils/log"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	engine *sql.DB

	stmtSysData  *sql.Stmt
	stmtSysMutex = utils.NewSemaphore(4)

	stmtAccount      *sql.Stmt
	stmtAccountMutex sync.Mutex

	stmtAccountActors      *sql.Stmt
	stmtAccountActorsMutex sync.Mutex

	stmtActorCount *sql.Stmt

	stmtInsertActor *sql.Stmt

	stmtQueryActor      *sql.Stmt
	stmtQueryCacheActor *sql.Stmt

	stmtUpdateActor      *sql.Stmt
	stmtUpdateActorMutex = utils.NewSemaphore(4)
)

func InitEngine() {
	var err error
	if engine, err = sql.Open("mysql", gconfig.GameConfig.Database); err != nil {
		panic(err)
	} else if err = engine.Ping(); err != nil {
		panic(err)
	}

	stmtSysData, err = engine.Prepare("update sysdata set data = ? where id = ?")
	if err != nil {
		panic(err)
	}

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

	stmtUpdateActor, err = engine.Prepare("update actor set actorname=?,level=?,power=?,logintime=?,logouttime=?,basedata=?,exdata=? where actorid=?")
	if err != nil {
		panic(err)
	}
}

func GetMaxActorId() (int64, error) {
	var maxId sql.NullInt64
	err := engine.QueryRow("select max(actorid) from actor where serverid=?", gconfig.GameConfig.ServerId).Scan(&maxId)
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

func UpdateSystemData(index int, value string, flush chan bool) {
	go func() {
		defer func() {
			if flush != nil {
				flush <- true
			}
		}()

		stmtSysMutex.Require()
		defer stmtSysMutex.Release()

		if _, err := stmtSysData.Exec(value, index); err != nil {
			log.Errorf("update system data %d error: %s", index, err.Error())
		}
	}()
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

func GetAccountActors(accountId int) ([]*AccountActor, error) {
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

	actors := make([]*AccountActor, 0)
	for rows.Next() {
		actor := &AccountActor{}
		if err := rows.Scan(&actor.ActorId, &actor.ActorName, &actor.Camp, &actor.Sex, &actor.Level); err != nil {
			return nil, err
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

func InsertActor(actor *Actor) error {
	baseData, err := json.Marshal(actor.BaseData)
	if err != nil {
		return err
	}
	exData, err := json.Marshal(actor.ExData)
	if err != nil {
		return err
	}
	_, err = stmtInsertActor.Exec(actor.ActorId, actor.ActorName, actor.AccountId,
		gconfig.GameConfig.ServerId, actor.Camp, actor.Sex, actor.Level, actor.Power,
		actor.CreateTime, actor.LoginTime, actor.LogoutTime, string(baseData), string(exData))
	return err
}

func QueryActor(actorId int64) (*Actor, error) {
	actor := &Actor{ActorId: actorId}
	var (
		baseData sql.NullString
		exData   sql.NullString
	)
	err := stmtQueryActor.QueryRow(actorId).Scan(&actor.ActorName, &actor.AccountId, &actor.Camp,
		&actor.Sex, &actor.Level, &actor.Power, &actor.LoginTime, &actor.LogoutTime, &baseData, &exData)
	if err != nil {
		return nil, err
	}
	actor.BaseData = &ActorBaseData{}
	actor.ExData = &ActorExData{}
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

func QueryActorCache(actorId int64) (*Actor, error) {
	actor := &Actor{ActorId: actorId}
	var (
		baseData sql.NullString
	)
	err := stmtQueryActor.QueryRow(actorId).Scan(&actor.ActorName, &actor.AccountId, &actor.Camp,
		&actor.Sex, &actor.Level, &actor.Power, &actor.LoginTime, &actor.LogoutTime, &baseData)
	if err != nil {
		return nil, err
	}
	actor.BaseData = &ActorBaseData{}
	if baseData.Valid {
		if err = json.Unmarshal([]byte(baseData.String), actor.BaseData); err != nil {
			return nil, err
		}
	}
	return actor, nil
}

func UpdateActor(actor *Actor, flush chan bool) {
	baseData, err := json.Marshal(actor.BaseData)
	if err != nil {
		log.Errorf("marshal actor(%d) base data error: %s", actor.ActorId, err.Error())
		return
	}

	exData, err := json.Marshal(actor.ExData)
	if err != nil {
		log.Errorf("marshal actor(%d) ex data error: %s", actor.ActorId, err.Error())
	}

	actorId, actorName, level, power, logintime, logouttime := actor.ActorId, actor.ActorName, actor.Level, actor.Power, actor.LoginTime, actor.LogoutTime
	go func() {
		defer func() {
			if flush != nil {
				flush <- true
			}
		}()
		stmtUpdateActorMutex.Require()
		defer stmtUpdateActorMutex.Release()

		_, err = stmtUpdateActor.Exec(actorName, level, power, logintime, logouttime, string(baseData), string(exData), actorId)
		if err != nil {
			log.Errorf("update actor(%d) error: %s", actorId, err.Error())
		}
	}()
}
